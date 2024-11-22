package view

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/view/event"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/panel"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const title = "CLIo"

// Main is the main view for the TUI.
type Main struct {
	ctx            context.Context
	commandManager manager

	keys   ckey.Map
	logger *slog.Logger

	// views
	searchPanel   panel.SearchView
	explorerPanel panel.ExplorerPanel
	detailPanel   panel.DetailsPanel
	executePanel  panel.ExecutePanel
	editPanel     panel.EditPanel
	explainPanel  panel.ExplainPanel
	help          help.Model

	focus     focus
	searching bool

	// styles
	titleStyle lipgloss.Style

	// Output is the view output
	Output string
}

// New returns a new main view.
func New(ctx context.Context, manager manager, logger *slog.Logger) (*Main, error) {
	m := Main{
		ctx:            ctx,
		commandManager: manager,
		titleStyle:     style.Title,
		keys:           ckey.DefaultMap,
		explorerPanel:  panel.NewExplorerPanel(),
		searchPanel:    panel.NewSearchView(logger),
		detailPanel:    panel.NewDetailsPanel(logger),
		executePanel:   panel.NewExecutePanel(logger),
		editPanel:      panel.NewEditPanel(logger),
		explainPanel:   panel.NewExplainPanel(logger),
		help:           help.New(),
		focus:          navigationFocus,
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

func (m *Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		// exit the app
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}
	// window resize
	case tea.WindowSizeMsg:
		hor, ver := style.Document.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-hor, msg.Height-ver)
		return m, nil
	// mode update
	case updateFocusMsg:
		msg.UpdateFocus(m)
		return m, m.initFocusedPanel()
	// handle outcome
	case event.ExecuteCommandMsg:
		m.Output = msg.Command
		return m, tea.Quit
	case event.NewCommandMsg:
		if err := m.saveCommand(msg.Command); err != nil {
			m.logger.Error("error storing new command", slog.Any("error", err))
		}
		return m, changeFocus(navigationFocus, nil)
	case event.UpdateCommandMsg:
		if err := m.editCommand(msg.Command); err != nil {
			m.logger.Error("error editing command", slog.Any("error", err))
		}
		return m, changeFocus(navigationFocus, nil)
	}
	return m, m.handleInput(msg)
}

func (m *Main) View() string {
	return style.Document.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			style.Container.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					// 1st column
					style.Border.BorderRight(true).Render(
						lipgloss.JoinVertical(
							lipgloss.Top,
							m.searchPanel.View(),
							style.Container.Render(m.explorerPanel.View()),
						),
					),
					// 2nd column
					lipgloss.JoinVertical(
						lipgloss.Center,
						m.titleStyle.Render(title),
						m.getPanelView(),
					)),
			),
			style.Help.Render(m.help.View(m.keys)),
		),
	)
}

func (m *Main) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return tea.Batch(
		m.explainPanel.Init(),
	)
}

func (m *Main) updateComponentsDimensions(width, height int) {
	// help
	m.help.Width = width

	// explorer panel
	w, h := util.RelativeDimensions(width, height, .20, .85)
	m.explorerPanel.SetSize(w, h)

	// search panel
	m.searchPanel.SetWidth(w)

	w, h = util.RelativeDimensions(width, height, .72, .91)
	// title
	m.titleStyle = m.titleStyle.Width(w)
	m.detailPanel.SetSize(w, h)
	m.executePanel.SetSize(w, h)
	m.editPanel.SetSize(w, h)
	m.explainPanel.SetSize(w, h)
}

func (m *Main) setContent(cmds []command.Command) error {
	if len(cmds) > 0 {
		cmd, err := m.fechFullCommand(cmds[0].ID.String())
		if err != nil {
			return err
		}
		cmds[0] = cmd
		m.detailPanel.SetCommand(cmd)
	}

	m.explorerPanel.SetCommands(cmds)
	return nil
}

func (m *Main) initFocusedPanel() tea.Cmd {
	switch m.focus {
	case searchFocus:
		return m.searchPanel.Init()
	case executeFocus:
		return m.executePanel.Init()
	case editFocus:
		return m.editPanel.Init()
	case explainFocus:
		return m.explainPanel.Init()
	}
	return nil
}

func (m *Main) getPanelView() string {
	switch m.focus {
	case executeFocus:
		return m.executePanel.View()
	case editFocus:
		return m.editPanel.View()
	case explainFocus:
		return m.explainPanel.View()
	default:
		return m.detailPanel.View()
	}
}
