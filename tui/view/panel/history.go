package panel

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	btable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

type History struct {
	logger       *slog.Logger
	keyMap       ckey.Map
	infoTable    *table.Table
	spinner      spinner.Model
	historyTable btable.Model

	loading   bool
	commandID uuid.UUID

	height       int
	width        int
	contentStyle lipgloss.Style
	titleStyle   lipgloss.Style
}

func NewHistory(keys ckey.Map, logger *slog.Logger) History {
	infoTable := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()

			if col != 0 {
				style = style.MarginLeft(1)
			}

			return style
		})

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	columns := []btable.Column{
		{Title: "Usage", Width: 32},
		{Title: "Timestamp", Width: 19},
	}

	t := btable.New(
		btable.WithColumns(columns),
		btable.WithHeight(7),
	)
	t.SetStyles(getTableStyles())

	return History{
		logger:       logger,
		keyMap:       keys,
		infoTable:    infoTable,
		historyTable: t,
		spinner:      s,
		titleStyle:   style.Title,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

func (p *History) Init() tea.Cmd {
	return p.spinner.Tick
}

func (p History) View() string {
	sty := lipgloss.NewStyle()
	cont := "Loading " + p.spinner.View()
	if !p.loading {
		cont = p.historyTable.View()
	}

	w := p.width - p.contentStyle.GetHorizontalBorderSize()
	h := p.height - p.contentStyle.GetVerticalFrameSize()

	return style.Border.Render(
		p.contentStyle.
			Width(w).
			Height(h).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					p.titleStyle.Render("History"),
					p.infoTable.Render(),
					sty.PaddingTop(1).
						Render(style.Label.Render("Usages")),
					cont,
				),
			))
}

func (p *History) Update(msg tea.Msg) (History, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keyMap.Go):
			if p.historyTable.Focused() {
				return *p, msgs.HandleExecuteMsg(p.commandID, p.historyTable.SelectedRow()[0])
			}
		default:
			p.historyTable, cmd = p.historyTable.Update(msg)
		}
	case spinner.TickMsg:
		p.spinner, cmd = p.spinner.Update(msg)
	}
	return *p, cmd
}

func (p *History) SetCommand(cmd command.Command) error {
	fmtCmd, err := util.FormatCommand(cmd.Command)
	if err != nil {
		return err
	}

	p.commandID = cmd.ID
	p.infoTable.Data(table.NewStringData([][]string{
		{style.Label.Render("Name"), style.Header.Render(cmd.Name)},
		{style.Label.Render("Description"), style.Header.Render(cmd.Description)},
		{style.Label.Render("Command"), style.Header.Render(fmtCmd)},
	}...))

	p.loading = true

	return nil
}

func (p *History) SetHistoryContent(history command.History) {
	p.loading = false

	rows := make([]btable.Row, 0, len(history.Usages))
	for _, usage := range history.Usages {
		rows = append(rows, btable.Row{
			usage.Command,
			usage.Timestamp.Format(time.RFC822),
		})
	}

	p.historyTable.SetRows(rows)
	p.historyTable.Focus()
	p.historyTable.SetCursor(0)
}

func (p *History) SetSize(width, height int) {
	p.height = height
	p.width = width

	p.titleStyle.Width(width)
	w, _ := util.RelativeDimensions(width, height, .6, .77)
	p.historyTable.Columns()[0].Width = w
	w, _ = util.RelativeDimensions(width, height, .2, .77)
	p.historyTable.Columns()[1].Width = w
}

func (p *History) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Back,
		p.historyTable.KeyMap.LineDown,
		p.historyTable.KeyMap.LineDown,
		p.keyMap.Go,
	}
}

func (p *History) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func getTableStyles() btable.Styles {
	s := btable.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("205")).
		Bold(false)

	return s
}
