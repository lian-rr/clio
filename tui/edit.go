package tui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian_rr/keep/command"
)

type editView struct {
	infoTable   *table.Table
	paramsTable *table.Table
	logger      *slog.Logger
	inputs      []*textinput.Model

	width         int
	height        int
	selectedInput int

	// styles
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style
	inputStyle   lipgloss.Style
}

func newEditView(logger *slog.Logger) editView {
	capitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	nameInput := textinput.New()
	nameInput.Placeholder = "Enter the command name"
	descInput := textinput.New()
	descInput.Placeholder = "and some description"
	cmdInput := textinput.New()
	cmdInput.Placeholder = "here goes the important part"

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

	return editView{
		infoTable:   infoTable,
		paramsTable: params,
		inputs:      []*textinput.Model{&nameInput, &descInput, &cmdInput},
		logger:      logger,
		titleStyle:  titleStyle,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
		inputStyle: lipgloss.NewStyle(),
	}
}

func (e *editView) Update(msg tea.KeyMsg) (editView, tea.Cmd) {
	inputCount := len(e.inputs)
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, defaultKeyMap.nextParamKey):
		e.inputs[e.selectedInput].Blur()
		e.selectedInput = (e.selectedInput + 1) % inputCount
		e.inputs[e.selectedInput].Focus()
	case key.Matches(msg, defaultKeyMap.previousParamKey):
		e.inputs[e.selectedInput].Blur()
		// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
		e.selectedInput = ((e.selectedInput-1)%inputCount + inputCount) % inputCount
		e.inputs[e.selectedInput].Focus()

	default:
		var input textinput.Model
		e.logger.Debug("update", slog.Int("selected input", e.selectedInput))
		input, cmd = e.inputs[e.selectedInput].Update(msg)
		e.inputs[e.selectedInput] = &input

	}
	return *e, cmd
}

func (v *editView) View() string {
	w := v.width - v.contentStyle.GetHorizontalBorderSize()
	h := v.height - v.contentStyle.GetVerticalFrameSize()

	style := lipgloss.NewStyle()
	v.infoTable.Data(table.NewStringData([][]string{
		{labelStyle.Render("Name"), v.inputStyle.Render(v.inputs[0].View())},
		{labelStyle.Render("Description"), v.inputStyle.Render(v.inputs[1].View())},
		{labelStyle.Render("Command"), v.inputStyle.Render(v.inputs[2].View())},
	}...))

	return borderStyle.Render(v.contentStyle.
		Width(w).
		Height(h).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				v.titleStyle.Render("New Command"),
				v.infoTable.Render(),
				style.MarginLeft(1).Render(labelStyle.Render("Parameters")),
				style.MarginLeft(2).Render(v.paramsTable.Render()),
			),
		))
}

func (v *editView) SetCommand(cmd *command.Command) error {
	v.logger.Debug("setting edit view content")
	if cmd == nil {
		v.inputs[0].Reset()
		v.inputs[1].Reset()
		v.inputs[2].Reset()
	}

	v.inputs[0].Focus()
	v.selectedInput = 0
	return nil
}

func (v *editView) SetSize(width, height int) {
	v.width = width
	v.height = height
	w, _ := relativeDimensions(width, height, .7, .7)
	v.infoTable.Width(w)
	v.paramsTable.Width(w)
	v.inputStyle = v.inputStyle.Width(w)
}
