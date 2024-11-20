package panel

import (
	"bytes"
	"log/slog"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const (
	chromaLang      = "fish"
	chromaFormatter = "terminal16m"
	chromaStyle     = "catppuccin-frappe"
)

// DetailsPanel handles the panel for showing the command details.
type DetailsPanel struct {
	view        viewport.Model
	infoTable   *table.Table
	paramsTable *table.Table
	logger      *slog.Logger

	// styles
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style
}

// NewDetailsPanel returns a new DetailsPanel.
func NewDetailsPanel(logger *slog.Logger) DetailsPanel {
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

	return DetailsPanel{
		infoTable:   infoTable,
		paramsTable: params,
		view:        viewport.New(0, 0),
		logger:      logger,
		titleStyle:  style.TitleStyle,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

// SetCommand sets the command to view in the panel.
func (p *DetailsPanel) SetCommand(cmd command.Command) error {
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
		{style.LabelStyle.Render("Name"), style.HeaderStyle.Render(cmd.Name)},
		{style.LabelStyle.Render("Description"), style.HeaderStyle.Render(cmd.Description)},
		{style.LabelStyle.Render("Command"), style.HeaderStyle.Render(b.String())},
	}...))

	sty := lipgloss.NewStyle()
	content := p.contentStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			p.infoTable.Render(),
			sty.MarginLeft(1).Render(style.LabelStyle.Render("Parameters")),
			sty.MarginLeft(2).Render(p.paramsTable.Render()),
		),
	)

	p.view.SetContent(content)
	return nil
}

func (p *DetailsPanel) View() string {
	return p.view.View()
}

func (p *DetailsPanel) SetSize(width, height int) {
	p.titleStyle.Width(width)

	p.view.Width, p.view.Height = width, height
	w, _ := util.RelativeDimensions(width, height, .90, .80)

	p.paramsTable.Width(w)
}
