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

type main struct {
	ctx            context.Context
	commandManager manager

	keys   keyMap
	logger *slog.Logger

	// views
	searchView   searchView
	commandsView listView
	detailView   detailsView
	executeView  executeView
	editView     editView
	help         help.Model

	currentMode mode
	searching   bool

	// styles
	titleStyle lipgloss.Style

	output string
}

func newMain(ctx context.Context, manager manager, logger *slog.Logger) (*main, error) {
	m := main{
		ctx:            ctx,
		commandManager: manager,
		titleStyle:     titleStyle,
		keys:           defaultKeyMap,
		commandsView:   newListView(),
		searchView:     newSearchView(),
		detailView:     newDetailsView(logger),
		executeView:    newExecuteView(logger),
		editView:       newEditView(logger),
		help:           help.New(),
		currentMode:    navigationMode,
		logger:         logger,
	}

	cmds, err := m.fechCommands()
	if err != nil {
		return nil, err
	}

	if err := m.setContent(cmds); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		hor, ver := docStyle.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-hor, msg.Height-ver)
		return m, nil
	// mode update
	case updateModeMsg:
		msg.updateMode(m)
		return m, nil
	// handle outcome
	case outcomeMsg:
		m.logger.Debug("output: ", slog.String("command", msg.output))
		m.output = msg.output
		return m, tea.Quit
	case newCmdMsg:
		m.logger.Debug("new command to store", slog.Any("command", msg.command))
		if err := m.saveCommand(msg.command); err != nil {
			m.logger.Error("error storing new command", slog.Any("error", err))
		}
		return m, changeMode(navigationMode, nil)
	case editCmdMsg:
		m.logger.Debug("command to edit", slog.Any("command", msg.command))
		return m, changeMode(navigationMode, nil)
	}
	return m, nil
}

func (m *main) View() string {
	var detailPanelContent string
	switch m.currentMode {
	case executeMode:
		detailPanelContent = m.executeView.View()
	case editMode:
		detailPanelContent = m.editView.View()
	default:
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

func (m *main) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return nil
}

func (m *main) updateComponentsDimensions(width, height int) {
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

	w, h = relativeDimensions(width, height, .74, .92)

	// execute view
	m.executeView.SetSize(w, h)

	// edit view
	m.editView.SetSize(w, h)
}

func (m *main) inputRouter(msg tea.KeyMsg) tea.Cmd {
	switch m.currentMode {
	case searchMode:
		return m.handleSearchInput(msg)
	case detailMode:
		return m.handleDetailInput(msg)
	case executeMode:
		return m.handleExecuteInput(msg)
	case editMode:
		return m.handleEditInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}
}

func (m *main) handleNavigationInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.quit):
		return tea.Quit
	case key.Matches(msg, m.keys.search):
		return changeMode(searchMode, func(m *main) {
			m.searchView.Focus()
		})
	case key.Matches(msg, m.keys.discardSearch):
		m.searchView.Reset()
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands", slog.Any("error", err))
			break
		}

		m.setContent(cmds)
	case key.Matches(msg, m.keys.enter):
		item, ok := m.commandsView.selectedItem()
		if !ok {
			break
		}

		return changeMode(executeMode, func(m *main) {
			err := m.executeView.SetCommand(*item.cmd)
			if err != nil {
				m.logger.Error("error setting execute view content", slog.Any("error", err))
			}
		})
	case key.Matches(msg, m.keys.new):
		return changeMode(editMode, func(m *main) {
			err := m.editView.SetCommand(newCommandMode, nil)
			if err != nil {
				m.logger.Error("error setting edit view content", slog.Any("error", err))
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

func (m *main) handleDetailInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, func(m *main) {
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

func (m *main) handleExecuteInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, func(m *main) {
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

func (m *main) handleEditInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, func(m *main) {
			item, ok := m.commandsView.selectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.editView, cmd = m.editView.Update(msg)
		return cmd
	}
}

func (m *main) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
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

func (m *main) setContent(cmds []command.Command) error {
	if len(cmds) > 0 {
		cmd, err := m.fechFullCommand(cmds[0].ID.String())
		if err != nil {
			return err
		}
		cmds[0] = cmd
		m.detailView.SetCommand(cmd)
	}

	m.commandsView.SetContent(cmds)
	return nil
}

func (m *main) fechCommands() ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.GetAll(ctx)
}

func (m *main) searchCommands(term string) ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.Search(ctx, term)
}

func (m *main) fechFullCommand(id string) (command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	return m.commandManager.GetOne(ctx, id)
}

func (m *main) saveCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	cmd, err := m.commandManager.Add(ctx, cmd)
	if err != nil {
		return err
	}

	idx := m.commandsView.AddItem(cmd)
	m.commandsView.Select(idx)
	m.detailView.SetCommand(cmd)

	return nil
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
