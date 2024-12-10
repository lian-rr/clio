package panel

import (
	"bytes"
	"log/slog"

	"github.com/alecthomas/chroma/v2/quick"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/components/dialog"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const (
	chromaLang      = "fish"
	chromaFormatter = "terminal16m"
	chromaStyle     = "catppuccin-frappe"
)

// Details handles the panel for showing the command details.
type Details struct {
	infoTable    *table.Table
	paramsTable  *table.Table
	confirmation dialog.Dialog
	logger       *slog.Logger

	width   int
	height  int
	confirm bool

	// styles
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style
}

// NewDetails returns a new DetailsPanel.
func NewDetails(logger *slog.Logger) Details {
	infoTable := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()

			if col != 0 {
				style = style.MarginLeft(1)
			}

			return style
		})

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("NAME", "DESCRIPTION", "DEFAULT VALUE")

	return Details{
		logger:       logger,
		infoTable:    infoTable,
		confirmation: dialog.New("Are you sure you want to delete the command?"),
		paramsTable:  params,
		titleStyle:   style.Title,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

// Update handles the msgs.
func (p *Details) Update(msg tea.Msg) (Details, tea.Cmd) {
	var cmd tea.Cmd
	p.confirmation, cmd = p.confirmation.Update(msg)
	return *p, cmd
}

// SetCommand sets the command to view in the panel.
func (p *Details) SetCommand(cmd command.Command) error {
	var b bytes.Buffer
	if err := quick.Highlight(&b, cmd.Command, chromaLang, chromaFormatter, chromaStyle); err != nil {
		return err
	}

	rows := make([][]string, 0, len(cmd.Params))
	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})
	}

	p.paramsTable.Data(table.NewStringData(rows...))

	p.infoTable.Data(table.NewStringData([][]string{
		{style.Label.Render("Name"), style.Header.Render(cmd.Name)},
		{style.Label.Render("Description"), style.Header.Render(cmd.Description)},
		{style.Label.Render("Command"), style.Header.Render(b.String())},
	}...))

	return nil
}

// View renders the DetailsPanel view.
func (p *Details) View() string {
	w := p.width - p.contentStyle.GetHorizontalBorderSize()
	h := p.height - p.contentStyle.GetVerticalFrameSize()

	var confirmation string
	if p.confirm {
		confirmation = p.confirmation.View()
	}

	sty := lipgloss.NewStyle()
	return style.Border.Render(
		p.contentStyle.
			Width(w).
			Height(h).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					p.titleStyle.Render("Details"),
					p.infoTable.Render(),
					confirmation,
					sty.MarginLeft(1).Render(style.Label.Render("Parameters")),
					p.paramsTable.Render(),
				),
			))
}

// SetSize sets the details panel size
func (p *Details) SetSize(width, height int) {
	p.titleStyle.Width(width)
	p.width = width
	p.height = height
	w, _ := util.RelativeDimensions(width, height, .7, .7)
	p.paramsTable.Width(w)
}

// ToggleConfirmation toggles the confirmation mode
func (p *Details) ToggleConfirmation() tea.Cmd {
	var cmd tea.Cmd
	if !p.confirm {
		cmd = p.confirmation.Init()
	}
	p.confirmation = p.confirmation.Reset()
	p.confirm = !p.confirm

	return cmd
}
