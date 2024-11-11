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

type view struct {
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

	output string
}

func newView(ctx context.Context, manager manager, logger *slog.Logger) (*view, error) {
	vi := view{
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

	cmds, err := vi.fechCommands()
	if err != nil {
		return nil, err
	}

	if err := vi.setContent(cmds); err != nil {
		return nil, err
	}

	return &vi, nil
}

func (v *view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		// exit the app
		if key.Matches(msg, v.keys.forceQuit) {
			return v, tea.Quit
		}
		return v, v.inputRouter(msg)
	// window resize
	case tea.WindowSizeMsg:
		hor, ver := docStyle.GetFrameSize()
		v.updateComponentsDimensions(msg.Width-hor, msg.Height-ver)
		return v, nil
	// mode update
	case updateModeMsg:
		msg.updateMode(v)
		return v, nil
	// handle outcome
	case outcomeMsg:
		v.logger.Debug("outcome", slog.String("command", msg.outcome))
		v.output = msg.outcome
		return v, tea.Quit
	}
	return v, nil
}

func (v *view) View() string {
	var detailPanelContent string
	if v.currentMode == executeMode {
		detailPanelContent = v.executeView.View()
	} else {
		detailPanelContent = v.detailView.View()
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
							v.searchView.View(),
							containerStyle.Render(v.commandsView.View()),
						),
					),
					// 2nd column
					lipgloss.JoinVertical(
						lipgloss.Top,
						v.titleStyle.Render(title),
						detailPanelContent,
					)),
			),
			helpStyle.Render(v.help.View(v.keys)),
		),
	)
}

func (v *view) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return nil
}

func (v *view) updateComponentsDimensions(width, height int) {
	// help
	v.help.Width = width

	// command explorer
	w, h := relativeDimensions(width, height, .20, .85)
	v.commandsView.SetSize(w, h)

	// search bar
	v.searchView.SetWidth(w)

	w, h = relativeDimensions(width, height, .75, .85)
	// title
	v.titleStyle = v.titleStyle.Width(w)

	// detail view
	v.detailView.SetSize(w, h)

	// execute view
	v.executeView.SetSize(w, h)

	v.logger.Debug("execute view",
		slog.Int("width", width),
		slog.Int("rel width", w),
		slog.Int("height", height),
		slog.Int("rel height", h),
	)
}

func (v *view) inputRouter(msg tea.KeyMsg) tea.Cmd {
	switch v.currentMode {
	case searchMode:
		return v.handleSearchInput(msg)
	case detailMode:
		return v.handleDetailInput(msg)
	case executeMode:
		return v.handleExecuteInput(msg)
	default:
		return v.handleNavigationInput(msg)
	}
}

func (v *view) handleNavigationInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, v.keys.quit):
		return tea.Quit
	case key.Matches(msg, v.keys.search):
		return changeMode(searchMode, func(m *view) {
			m.searchView.Focus()
		})
	case key.Matches(msg, v.keys.discardSearch):
		v.searchView.Reset()
		cmds, err := v.fechCommands()
		if err != nil {
			v.logger.Error("error getting all commands",
				slog.Any("error", err),
			)
			break
		}

		v.setContent(cmds)
	case key.Matches(msg, v.keys.enter):
		item, ok := v.commandsView.selectedItem()
		if !ok {
			break
		}

		return changeMode(executeMode, func(m *view) {
			err := m.executeView.SetCommand(*item.cmd)
			if err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		v.commandsView, cmd = v.commandsView.Update(msg)
		item, ok := v.commandsView.selectedItem()
		if !ok {
			break
		}

		if !item.loaded {
			c, err := v.commandManager.GetOne(v.ctx, item.cmd.ID.String())
			if err != nil {
				v.logger.Error("error fetching command details", slog.Any("error", err))
				break
			}

			item.cmd.Params = c.Params
			item.loaded = true

			v.logger.Debug("command details fetched successfully", slog.Any("command", c))
		}
		if err := v.detailView.SetCommand(*item.cmd); err != nil {
			v.logger.Error("error setting detail view content", slog.Any("error", err))
		}
	}
	return cmd
}

func (v *view) handleDetailInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, v.keys.back):
		return changeMode(navigationMode, func(m *view) {
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

func (v *view) handleExecuteInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, v.keys.back):
		return changeMode(navigationMode, func(m *view) {
			item, ok := m.commandsView.selectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		v.executeView, cmd = v.executeView.Update(msg)
		return cmd
	}
}

func (v *view) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	getAll := func() {
		cmds, err := v.fechCommands()
		if err != nil {
			v.logger.Error("error getting all commands", slog.Any("error", err))
			return
		}

		v.setContent(cmds)
	}

	var cmd tea.Cmd
	switch {
	case key.Matches(msg, v.keys.back):
		return changeMode(navigationMode, nil)
	case key.Matches(msg, v.keys.discardSearch):
		v.searchView.Reset()
		getAll()
		return changeMode(navigationMode, nil)
	case key.Matches(msg, v.keys.enter):
		v.searchView.Unfocus()
		return changeMode(navigationMode, nil)
	default:
		v.searchView, cmd = v.searchView.Update(msg)

		terms := v.searchView.Content()
		if len(terms) >= minCharCount {
			v.searching = true
			cmds, err := v.searchCommands(terms)
			if err != nil {
				v.logger.Error("error searching for commands",
					slog.String("terms", terms),
					slog.Any("error", err),
				)
				break
			}
			v.setContent(cmds)
		} else if v.searching && len(terms) == 0 {
			v.searching = false
			getAll()
		}
	}
	return cmd
}

func (v *view) setContent(cmds []command.Command) error {
	if len(cmds) > 0 {
		cmd, err := v.fechFullCommand(cmds[0].ID.String())
		if err != nil {
			return err
		}
		cmds[0] = cmd
		v.detailView.SetCommand(cmd)
		v.logger.Debug("default cmd", cmds[0])
	}

	v.commandsView.SetContent(cmds)
	return nil
}

func (v *view) fechCommands() ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(v.ctx, time.Millisecond*300)
	defer cancel()

	return v.commandManager.GetAll(ctx)
}

func (v *view) searchCommands(term string) ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(v.ctx, time.Millisecond*300)
	defer cancel()

	return v.commandManager.Search(ctx, term)
}

func (v *view) fechFullCommand(id string) (command.Command, error) {
	ctx, cancel := context.WithTimeout(v.ctx, time.Millisecond*200)
	defer cancel()

	return v.commandManager.GetOne(ctx, id)
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
