package panel

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian-rr/clio/command"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/mode"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const (
	nameInputPos = iota
	descInputPos
	cmdInputPos
)

type cmdEditMode int

const (
	_ cmdEditMode = iota
	NewCommandMode
	EditCommandMode
)

// number of fixed inputs (name, description, command)
const fixedInputs = 3

type EditView struct {
	cmd  *command.Command
	mode cmdEditMode
	// cache the params inputs
	paramsContent map[string][2]*textinput.Model

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

func NewEditView(logger *slog.Logger) EditView {
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

	return EditView{
		mode:          NewCommandMode,
		infoTable:     infoTable,
		paramsTable:   params,
		inputs:        []*textinput.Model{&nameInput, &descInput, &cmdInput},
		paramsContent: make(map[string][2]*textinput.Model),
		logger:        logger,
		titleStyle:    style.TitleStyle,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
		inputStyle: lipgloss.NewStyle(),
	}
}

func (v *EditView) Update(msg tea.KeyMsg) (EditView, tea.Cmd) {
	inputCount := len(v.inputs)
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, ckey.DefaultMap.NextParamKey):
		v.inputs[v.selectedInput].Blur()
		v.selectedInput = (v.selectedInput + 1) % inputCount
		v.inputs[v.selectedInput].Focus()
	case key.Matches(msg, ckey.DefaultMap.PreviousParamKey):
		v.inputs[v.selectedInput].Blur()
		// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
		v.selectedInput = ((v.selectedInput-1)%inputCount + inputCount) % inputCount
		v.inputs[v.selectedInput].Focus()
	case key.Matches(msg, ckey.DefaultMap.Enter):
		if err := v.cmd.Build(); err != nil {
			v.logger.Warn("error building param", slog.Any("error", err))
			break
		}
		switch v.mode {
		case NewCommandMode:
			return *v, mode.HandleNewCmdMsg(*v.cmd)
		case EditCommandMode:
			return *v, mode.HandleUpdateCmd(*v.cmd)
		default:
			v.logger.Error("unknown mode found. discarding command", slog.Any("mode", v.mode))
		}
	default:
		var input textinput.Model
		input, cmd = v.inputs[v.selectedInput].Update(msg)
		v.inputs[v.selectedInput] = &input

		// if command didn't changed
		if v.selectedInput > cmdInputPos {
			v.updateParams()
		} else {
			if err := v.updateCommand(); err != nil {
				v.logger.Warn("error building cmd", slog.Any("error", err))
			}
		}

		v.logger.Debug("cmd values updated",
			slog.String("name", v.cmd.Name),
			slog.String("desc", v.cmd.Description),
			slog.String("command", v.cmd.Command),
			slog.Any("params", v.cmd.Params),
		)
	}
	return *v, cmd
}

func (v *EditView) View() string {
	w := v.width - v.contentStyle.GetHorizontalBorderSize()
	h := v.height - v.contentStyle.GetVerticalFrameSize()

	v.infoTable.Data(table.NewStringData([][]string{
		{style.LabelStyle.Render("Name"), v.inputStyle.Render(v.inputs[nameInputPos].View())},
		{style.LabelStyle.Render("Description"), v.inputStyle.Render(v.inputs[descInputPos].View())},
		{style.LabelStyle.Render("Command"), v.inputStyle.Render(v.inputs[cmdInputPos].View())},
	}...))

	rows := make([][]string, 0, len(v.cmd.Params))
	for i, param := range v.cmd.Params {
		rows = append(rows, []string{
			param.Name,
			v.inputs[i*2+3].View(),
			v.inputs[i*2+4].View(),
		})
	}

	v.paramsTable.Data(table.NewStringData(rows...))

	var title string
	switch v.mode {
	case EditCommandMode:
		title = "Edit Command"
	default:
		title = "New Command"
	}

	sty := lipgloss.NewStyle()
	return style.BorderStyle.Render(v.contentStyle.
		Width(w).
		Height(h).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				v.titleStyle.Render(title),
				v.infoTable.Render(),
				sty.MarginLeft(1).Render(style.LabelStyle.Render("Parameters")),
				sty.MarginLeft(2).Render(v.paramsTable.Render()),
			),
		))
}

func (v *EditView) SetCommand(mode cmdEditMode, cmd *command.Command) error {
	// clear the params inputs
	for _, input := range v.inputs {
		input.Reset()
	}
	v.inputs = append([]*textinput.Model{}, v.inputs[:fixedInputs]...)
	v.paramsContent = make(map[string][2]*textinput.Model)

	v.mode = mode
	if cmd == nil {
		v.cmd = &command.Command{}
	} else {
		v.cmd = cmd
		v.inputs[nameInputPos].SetValue(cmd.Name)
		v.inputs[descInputPos].SetValue(cmd.Description)
		v.inputs[cmdInputPos].SetValue(cmd.Command)
		v.refreshParamsInputs()
	}

	v.inputs[nameInputPos].Focus()
	v.selectedInput = nameInputPos
	return nil
}

func (v *EditView) SetSize(width, height int) {
	v.width = width
	v.height = height
	w, _ := util.RelativeDimensions(width, height, .7, .7)
	v.infoTable.Width(w)
	v.paramsTable.Width(w)
	v.inputStyle = v.inputStyle.Width(w)
}

func (v *EditView) updateCommand() error {
	v.cmd.Name = v.inputs[nameInputPos].Value()
	v.cmd.Description = v.inputs[descInputPos].Value()

	cmd := v.inputs[cmdInputPos].Value()
	if len(cmd) != len(v.cmd.Command) {
		v.cmd.Command = v.inputs[cmdInputPos].Value()
		v.logger.Debug("rebuilding command")
		if err := v.cmd.Build(); err != nil {
			return err
		}
		v.refreshParamsInputs()
	}

	return nil
}

func (v *EditView) updateParams() {
	paramPos := (v.selectedInput - fixedInputs) / 2
	field := (v.selectedInput - fixedInputs) % 2

	pName := v.cmd.Params[paramPos].Name
	value := v.inputs[v.selectedInput].Value()
	v.paramsContent[pName][field].SetValue(value)

	if field == 0 {
		v.cmd.Params[paramPos].Description = value
	} else {
		v.cmd.Params[paramPos].DefaultValue = value
	}
}

func (v *EditView) refreshParamsInputs() {
	inputs := v.inputs[:fixedInputs]
	for _, param := range v.cmd.Params {
		if in, ok := v.paramsContent[param.Name]; ok {
			inputs = append(inputs, in[0], in[1])
		} else {
			descInput := textinput.New()
			descInput.Placeholder = "add some description"
			descInput.SetValue(param.Description)

			dvInput := textinput.New()
			dvInput.Placeholder = "optional"
			v.paramsContent[param.Name] = [2]*textinput.Model{&descInput, &dvInput}
			dvInput.SetValue(param.DefaultValue)

			inputs = append(inputs, &descInput, &dvInput)
		}
	}
	v.inputs = inputs
}
