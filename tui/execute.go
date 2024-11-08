package tui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian_rr/keep/command"
)

type executeView struct {
	command *command.Command

	paramsTable   *table.Table
	infoTable     *table.Table
	paramInputs   map[string]*textinput.Model
	selectedInput string

	width  int
	height int

	contentStyle lipgloss.Style
	titleStyle   lipgloss.Style
	logger       *slog.Logger
}

func newExecuteView(logger *slog.Logger) executeView {
	infoTable := table.New().Border(lipgloss.HiddenBorder())

	capitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	paramHeaders := []string{
		"name",
		"description",
		"default value",
	}

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(capitalizeHeaders(paramHeaders)...)

	return executeView{
		logger:      logger,
		infoTable:   infoTable,
		paramsTable: params,
		titleStyle: labelStyle.
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle),
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

func (v *executeView) Update(msg tea.KeyMsg) (executeView, tea.Cmd) {
	var cmd tea.Cmd
	if len(v.paramInputs) != 0 {
		var input textinput.Model
		input, cmd = v.paramInputs[v.selectedInput].Update(msg)
		v.paramInputs[v.selectedInput] = &input
	}
	return *v, cmd
}

func (v *executeView) View() string {
	if v.command == nil {
		return ""
	}

	arguments := make([]command.Argument, 0, len(v.command.Params))
	for param, input := range v.paramInputs {
		arguments = append(arguments, command.Argument{
			Name:  param,
			Value: input.View(),
		})
	}

	outCommand, err := v.command.Compile(arguments)
	if err != nil {
		v.logger.Error("error compiling command",
			slog.String("name", v.command.Name),
			slog.String("command", v.command.Command),
			slog.Any("params", v.command.Params),
			slog.Any("arguments", arguments),
			slog.Any("error", err),
		)
		return ""
	}

	w := v.width - v.contentStyle.GetHorizontalBorderSize()
	h := v.height - v.contentStyle.GetVerticalFrameSize()

	return borderStyle.Render(v.contentStyle.
		Width(w).
		Height(h).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				v.titleStyle.Render("Compose"),
				v.infoTable.Render(),
				outCommand,
				v.paramsTable.Render(),
			),
		))
}

func (v *executeView) SetCommand(cmd command.Command) error {
	v.command = &cmd

	v.infoTable.Data(table.NewStringData([][]string{
		{labelStyle.Render("Title"), cmd.Name},
		{labelStyle.Render("Description"), cmd.Description},
	}...))

	rows := make([][]string, 0, len(cmd.Params))
	v.paramInputs = make(map[string]*textinput.Model, len(cmd.Params))
	for i, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})

		pi := textinput.New()
		pi.Placeholder = param.Name
		pi.TextStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.AdaptiveColor{
				Light: "#909090",
				Dark:  "#626262",
			})
		pi.Prompt = ""
		if param.DefaultValue != "" {
			pi.SetSuggestions([]string{param.DefaultValue})
		}

		v.paramInputs[param.Name] = &pi
		if i == 0 {
			v.selectedInput = param.Name
			pi.Focus()
		}
	}

	v.paramsTable.Data(table.NewStringData(rows...))

	return nil
}

func (v *executeView) SetSize(width, height int) {
	v.width = width
	v.height = height
}
