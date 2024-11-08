package tui

import (
	"context"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian_rr/keep/command"
)

const title = "KEEP"

const minCharCount = 3

type model struct {
	ctx            context.Context
	commandManager manager

	keys   keyMap
	logger *slog.Logger

	// panels
	searchView   searchView
	commandsView listView
	detailView   detailsView
	executeView  executeView
	help         help.Model

	currentMode mode
	searching   bool

	// styles
	titleStyle lipgloss.Style
}

func newModel(ctx context.Context, manager manager, logger *slog.Logger) (*model, error) {
	model := model{
		ctx:            ctx,
		commandManager: manager,
		titleStyle:     titleStyle,
		keys:           defaultKeyMap,
		commandsView:   newListView(),
		searchView:     newSearchView(),
		detailView:     newDetailsView(logger),
		executeView:    newExecuteView(logger),
		help:           help.New(),
		currentMode:    navigationMode,
		logger:         logger,
	}

	cmds, err := model.fechCommands()
	if err != nil {
		return nil, err
	}

	if err := model.setContent(cmds); err != nil {
		return nil, err
	}

	return &model, nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		// exit the app
		if key.Matches(msg, m.keys.forceQuit) {
			return m, tea.Quit
		}
		return m, m.inputRouter(msg)
	// window resize
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-h, msg.Height-v)
		return m, nil
	// mode update
	case updateModeMsg:
		msg.updateMode(m)
		return m, nil
	}

	return m, nil
}

func (m *model) View() string {
	var detailPanelContent string
	if m.currentMode == executeMode {
		detailPanelContent = m.executeView.View()
	} else {
		detailPanelContent = m.detailView.View()
	}

	return docStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			containerStyle.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					// 1st column
					borderStyle.BorderRight(true).Render(
						lipgloss.JoinVertical(
							lipgloss.Top,
							m.searchView.View(),
							containerStyle.Render(m.commandsView.View()),
						),
					),
					// 2nd column
					lipgloss.JoinVertical(
						lipgloss.Top,
						m.titleStyle.Render(title),
						detailPanelContent,
					)),
			),
			helpStyle.Render(m.help.View(m.keys)),
		),
	)
}

func (m *model) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return nil
}

func (m *model) updateComponentsDimensions(width, height int) {
	// help
	m.help.Width = width

	// command explorer
	w, h := relativeDimensions(width, height, .20, .85)
	m.commandsView.SetSize(w, h)

	// search bar
	m.searchView.SetWidth(w)

	w, h = relativeDimensions(width, height, .75, .85)
	// title
	m.titleStyle = m.titleStyle.Width(w)

	// detail view
	m.detailView.SetSize(w, h)

	// execute view
	m.executeView.SetSize(w, h)

	m.logger.Debug("execute view",
		slog.Int("width", width),
		slog.Int("rel width", w),
		slog.Int("height", height),
		slog.Int("rel height", h),
	)
}

func (m *model) inputRouter(msg tea.KeyMsg) tea.Cmd {
	switch m.currentMode {
	case searchMode:
		return m.handleSearchInput(msg)
	case detailMode:
		return m.handleDetailInput(msg)
	case executeMode:
		return m.handleExecuteInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}
}

func (m *model) handleNavigationInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.quit):
		return tea.Quit
	case key.Matches(msg, m.keys.search):
		return changeMode(searchMode, func(m *model) {
			m.searchView.Focus()
		})
	case key.Matches(msg, m.keys.discardSearch):
		m.searchView.Reset()
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands",
				slog.Any("error", err),
			)
			break
		}

		m.setContent(cmds)
	case key.Matches(msg, m.keys.enter):
		item, ok := m.commandsView.selectedItem()
		if !ok {
			break
		}

		return changeMode(executeMode, func(m *model) {
			err := m.executeView.SetCommand(*item.cmd)
			if err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.commandsView, cmd = m.commandsView.Update(msg)
		item, ok := m.commandsView.selectedItem()
		if !ok {
			break
		}

		if !item.loaded {
			c, err := m.commandManager.GetOne(m.ctx, item.cmd.ID.String())
			if err != nil {
				m.logger.Error("error fetching command details", slog.Any("error", err))
				break
			}

			item.cmd.Params = c.Params
			item.loaded = true

			m.logger.Debug("command details fetched successfully", slog.Any("command", c))
		}
		if err := m.detailView.SetCommand(*item.cmd); err != nil {
			m.logger.Error("error setting detail view content", slog.Any("error", err))
		}
	}
	return cmd
}

func (m *model) handleDetailInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, func(m *model) {
			item, ok := m.commandsView.selectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		return cmd
	}
}

func (m *model) handleExecuteInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, func(m *model) {
			item, ok := m.commandsView.selectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.executeView, cmd = m.executeView.Update(msg)
		return cmd
	}
}

func (m *model) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	getAll := func() {
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands", slog.Any("error", err))
			return
		}

		m.setContent(cmds)
	}

	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, nil)
	case key.Matches(msg, m.keys.discardSearch):
		m.searchView.Reset()
		getAll()
		return changeMode(navigationMode, nil)
	case key.Matches(msg, m.keys.enter):
		m.searchView.Unfocus()
		return changeMode(navigationMode, nil)
	default:
		m.searchView, cmd = m.searchView.Update(msg)

		terms := m.searchView.Content()
		if len(terms) >= minCharCount {
			m.searching = true
			cmds, err := m.searchCommands(terms)
			if err != nil {
				m.logger.Error("error searching for commands",
					slog.String("terms", terms),
					slog.Any("error", err),
				)
				break
			}
			m.setContent(cmds)
		} else if m.searching && len(terms) == 0 {
			m.searching = false
			getAll()
		}
	}
	return cmd
}

func (m *model) setContent(cmds []command.Command) error {
	m.commandsView.SetContent(cmds)
	if len(cmds) > 0 {
		cmd, err := m.fechFullCommand(cmds[0].ID.String())
		if err != nil {
			return err
		}
		cmds[0] = cmd
		m.detailView.SetCommand(cmd)
	}

	return nil
}

func (m *model) fechCommands() ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.GetAll(ctx)
}

func (m *model) searchCommands(term string) ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.Search(ctx, term)
}

func (m *model) fechFullCommand(id string) (command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	return m.commandManager.GetOne(ctx, id)
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
