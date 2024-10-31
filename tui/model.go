package tui

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const title = "KEEP"

type model struct {
	ctx            context.Context
	commandManager manager

	keys   keyMap
	logger *slog.Logger

	// panels
	commands   listView
	detailView detailsView
	searchView searchView
	help       help.Model

	currentMode mode

	// styles
	titleStyle lipgloss.Style
}

func newModel(ctx context.Context, manager manager, logger *slog.Logger) (*model, error) {
	cmds, err := manager.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	detail := newDetailsView(logger)
	if len(cmds) > 0 {
		cmd, err := manager.GetOne(ctx, cmds[0].ID.String())
		if err != nil {
			return nil, err
		}
		cmds[0] = cmd
		detail.SetContent(cmds[0])
	}

	model := model{
		ctx:            ctx,
		commandManager: manager,
		titleStyle:     titleStyle,
		keys:           defaultKeyMap,
		commands:       newListView("Commands", cmds),
		detailView:     detail,
		searchView:     newSearchView(),
		help:           help.New(),
		currentMode:    navigationMode,
		logger:         logger,
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
							containerStyle.Render(m.commands.View()),
						),
					),
					// 2nd column
					lipgloss.JoinVertical(
						lipgloss.Top,
						m.titleStyle.Render(title),
						m.detailView.View(),
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
	m.commands.SetSize(w, h)

	// search bar
	m.searchView.SetWidth(w)

	w, h = relativeDimensions(width, height, .75, .85)
	// title
	m.titleStyle = m.titleStyle.Width(w)
	// detail view
	m.detailView.SetSize(w, h)
}

func (m *model) inputRouter(msg tea.KeyMsg) tea.Cmd {
	switch m.currentMode {
	case searchMode:
		return m.handleSearchInput(msg)
	case createMode:
	case editMode:
	case detailMode:
		return m.handleDetailInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}

	return nil
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
	case key.Matches(msg, m.keys.enter):
		command, err := m.commands.selectedItem()
		if err != nil {
			m.logger.Error("error getting selected command", slog.Any("error", err))
			break
		}

		return changeMode(detailMode, func(m *model) {
			err := m.detailView.SetContent(*command.cmd)
			if err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.commands, cmd = m.commands.Update(msg)
		command, err := m.commands.selectedItem()
		if err != nil {
			m.logger.Error("error getting selected item", slog.Any("error", err))
			break
		}

		if !command.loaded {
			c, err := m.commandManager.GetOne(m.ctx, command.cmd.ID.String())
			if err != nil {
				m.logger.Error("error fetching command details", slog.Any("error", err))
				break
			}

			command.cmd.Params = c.Params
			command.loaded = true

			m.logger.Debug("command details fetched successfully", slog.Any("command", c))
		}
		if err := m.detailView.SetContent(*command.cmd); err != nil {
			m.logger.Error("error setting detail view content", slog.Any("error", err))
		}
	}
	return cmd
}

func (m *model) handleDetailInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, nil)
	default:
		// NOTE: nothing for the moment
	}
	return cmd
}

func (m *model) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, nil)
	case key.Matches(msg, m.keys.enter):
		m.logger.Debug("search", slog.String("terms", m.searchView.Content()))
		// TODO: fetch the commands and load them in commands view.
		// TODO: think on how to get all the commands again
		return changeMode(navigationMode, nil)
	default:
		m.searchView, cmd = m.searchView.Update(msg)
	}
	return cmd
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
