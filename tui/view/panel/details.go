package panel

import (
	"bytes"
	"log/slog"
	"strings"

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

type DetailsView struct {
	view        viewport.Model
	infoTable   *table.Table
	paramsTable *table.Table
	logger      *slog.Logger

	// styles
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style
}

func NewDetailsView(logger *slog.Logger) DetailsView {
	capitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	infoTable := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()

			if col != 0 {
				style = style.MarginLeft(1)
			}

			return style
		})

	paramHeaders := []string{
		"name",
		"description",
		"default value",
	}

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(capitalizeHeaders(paramHeaders)...)

	return DetailsView{
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

func (dc *DetailsView) SetCommand(cmd command.Command) error {
	var b bytes.Buffer
	if err := quick.Highlight(&b, cmd.Command, chromaLang, chromaFormatter, chromaStyle); err != nil {
		return err
	}

	rows := make([][]string, 0, len(cmd.Params))
	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})
	}

	dc.paramsTable.Data(table.NewStringData(rows...))

	dc.infoTable.Data(table.NewStringData([][]string{
		{style.LabelStyle.Render("Name"), style.HeaderStyle.Render(cmd.Name)},
		{style.LabelStyle.Render("Description"), style.HeaderStyle.Render(cmd.Description)},
		{style.LabelStyle.Render("Command"), style.HeaderStyle.Render(b.String())},
	}...))

	sty := lipgloss.NewStyle()
	content := dc.contentStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			dc.infoTable.Render(),
			sty.MarginLeft(1).Render(style.LabelStyle.Render("Parameters")),
			sty.MarginLeft(2).Render(dc.paramsTable.Render()),
		),
	)

	dc.view.SetContent(content)
	return nil
}

func (dc *DetailsView) View() string {
	return dc.view.View()
}

func (dc *DetailsView) SetSize(width, height int) {
	dc.titleStyle.Width(width)

	dc.view.Width, dc.view.Height = width, height
	w, _ := util.RelativeDimensions(width, height, .90, .80)

	dc.paramsTable.Width(w)
}
