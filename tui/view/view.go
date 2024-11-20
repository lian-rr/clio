package view

import (
	"context"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

type manager interface {
	GetAll(context.Context) ([]command.Command, error)
	GetOne(context.Context, string) (command.Command, error)
	Search(context.Context, string) ([]command.Command, error)
	Add(context.Context, command.Command) (command.Command, error)
	DeleteCommand(context.Context, string) error
	UpdateCommand(context.Context, command.Command) (command.Command, error)
}

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
		titleStyle:     style.TitleStyle,
		keys:           ckey.DefaultMap,
		explorerPanel:  panel.NewExplorerPanel(),
		searchPanel:    panel.NewSearchView(),
		detailPanel:    panel.NewDetailsPanel(logger),
		executePanel:   panel.NewExecutePanel(logger),
		editPanel:      panel.NewEditPanel(logger),
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
		return m, m.handleInput(msg)
	// window resize
	case tea.WindowSizeMsg:
		hor, ver := style.DocStyle.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-hor, msg.Height-ver)
		return m, nil
	// mode update
	case updateFocusMsg:
		msg.UpdateFocus(m)
		return m, nil
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
	return m, nil
}

func (m *Main) View() string {
	var detailPanelContent string
	switch m.focus {
	case executeFocus:
		detailPanelContent = m.executePanel.View()
	case editFocus:
		detailPanelContent = m.editPanel.View()
	default:
		detailPanelContent = m.detailPanel.View()
	}

	return style.DocStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			style.ContainerStyle.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					// 1st column
					style.BorderStyle.BorderRight(true).Render(
						lipgloss.JoinVertical(
							lipgloss.Top,
							m.searchPanel.View(),
							style.ContainerStyle.Render(m.explorerPanel.View()),
						),
					),
					// 2nd column
					lipgloss.JoinVertical(
						lipgloss.Top,
						m.titleStyle.Render(title),
						detailPanelContent,
					)),
			),
			style.HelpStyle.Render(m.help.View(m.keys)),
		),
	)
}

func (m *Main) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return textinput.Blink
}

func (m *Main) updateComponentsDimensions(width, height int) {
	// help
	m.help.Width = width

	// explorer panel
	w, h := util.RelativeDimensions(width, height, .20, .85)
	m.explorerPanel.SetSize(w, h)

	// search panel
	m.searchPanel.SetWidth(w)

	w, h = util.RelativeDimensions(width, height, .75, .85)
	// title
	m.titleStyle = m.titleStyle.Width(w)

	// detail panel
	m.detailPanel.SetSize(w, h)

	w, h = util.RelativeDimensions(width, height, .74, .92)

	// execute panel
	m.executePanel.SetSize(w, h)

	// edit panel
	m.editPanel.SetSize(w, h)
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

func (m *Main) fechCommands() ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.GetAll(ctx)
}

func (m *Main) searchCommands(term string) ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandManager.Search(ctx, term)
}

func (m *Main) fechFullCommand(id string) (command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	return m.commandManager.GetOne(ctx, id)
}

func (m *Main) saveCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	cmd, err := m.commandManager.Add(ctx, cmd)
	if err != nil {
		return err
	}

	idx := m.explorerPanel.AddCommand(cmd)
	m.explorerPanel.Select(idx)
	m.detailPanel.SetCommand(cmd)

	return nil
}

func (m *Main) editCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	newCmd, err := m.commandManager.UpdateCommand(ctx, cmd)
	if err != nil {
		return err
	}

	m.explorerPanel.RefreshCommand(newCmd)
	m.detailPanel.SetCommand(newCmd)
	return nil
}

func (m *Main) removeCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	err := m.commandManager.DeleteCommand(ctx, cmd.ID.String())
	if err != nil {
		return err
	}

	toSelectPos := m.explorerPanel.RemoveSelectedCommand()
	if toSelectPos >= 0 {
		m.explorerPanel.Select(toSelectPos)
		if item, ok := m.explorerPanel.SelectedCommand(); ok {
			m.detailPanel.SetCommand(*item.Command)
		}
	} else {
		m.detailPanel.SetCommand(command.Command{})
	}

	return nil
}