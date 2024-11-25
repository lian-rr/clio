package panel

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian-rr/clio/command"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const (
	nameInputPos = iota
	descInputPos
	cmdInputPos
)

// EditPanelMode represents the way the panel is going to be used.
type EditPanelMode int

const (
	_ EditPanelMode = iota
	// NewCommandMode pannel is going to return a new Command
	NewCommandMode
	// EditCommandMode pannel is going to return the passed Command updated.
	EditCommandMode
)

// number of fixed inputs (name, description, command)
const fixedInputs = 3

// EditPanel handles the panel for editing or creating a command.
type EditPanel struct {
	cmd  command.Command
	mode EditPanelMode
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

// NewEditPanel returns a new ExecutePanel.
func NewEditPanel(logger *slog.Logger) EditPanel {
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

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("NAME", "DESCRIPTION", "DEFAULT VALUE")

	return EditPanel{
		mode:          NewCommandMode,
		infoTable:     infoTable,
		paramsTable:   params,
		inputs:        []*textinput.Model{&nameInput, &descInput, &cmdInput},
		paramsContent: make(map[string][2]*textinput.Model),
		logger:        logger,
		titleStyle:    style.Title,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
		inputStyle: lipgloss.NewStyle(),
	}
}

// Init starts the input blink
func (p *EditPanel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles the msgs.
func (p *EditPanel) Update(msg tea.Msg) (EditPanel, tea.Cmd) {
	inputCount := len(p.inputs)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, ckey.DefaultMap.NextParamKey):
			p.inputs[p.selectedInput].Blur()
			p.selectedInput = (p.selectedInput + 1) % inputCount
			p.inputs[p.selectedInput].Focus()
		case key.Matches(msg, ckey.DefaultMap.PreviousParamKey):
			p.inputs[p.selectedInput].Blur()
			// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
			p.selectedInput = ((p.selectedInput-1)%inputCount + inputCount) % inputCount
			p.inputs[p.selectedInput].Focus()
		case key.Matches(msg, ckey.DefaultMap.Enter):
			if err := p.cmd.Build(); err != nil {
				p.logger.Warn("error building param", slog.Any("error", err))
				break
			}

			p.logger.Debug("Done editing/creating command", slog.Any("command", p.cmd))
			switch p.mode {
			case NewCommandMode:
				return *p, msgs.HandleNewCommandMsg(p.cmd)
			case EditCommandMode:
				return *p, msgs.HandleUpdateCommandMsg(p.cmd)
			default:
				p.logger.Error("unknown mode found. discarding command", slog.Any("mode", p.mode))
			}
		default:
			var input textinput.Model
			input, cmd = p.inputs[p.selectedInput].Update(msg)
			p.inputs[p.selectedInput] = &input

			// command didn't changed
			if p.selectedInput > cmdInputPos {
				p.updateParams()
			} else {
				if err := p.updateCommand(); err != nil {
					p.logger.Warn("error building cmd", slog.Any("error", err))
				}
			}
		}
	default:
		var input textinput.Model
		input, cmd = p.inputs[p.selectedInput].Update(msg)
		p.inputs[p.selectedInput] = &input
	}
	return *p, cmd
}

// View returns the string representation of the panel.
func (p *EditPanel) View() string {
	w := p.width - p.contentStyle.GetHorizontalBorderSize()
	h := p.height - p.contentStyle.GetVerticalFrameSize()

	p.infoTable.Data(table.NewStringData([][]string{
		{style.Label.Render("Name"), p.inputStyle.Render(p.inputs[nameInputPos].View())},
		{style.Label.Render("Description"), p.inputStyle.Render(p.inputs[descInputPos].View())},
		{style.Label.Render("Command"), p.inputStyle.Render(p.inputs[cmdInputPos].View())},
	}...))

	rows := make([][]string, 0, len(p.cmd.Params))
	for i, param := range p.cmd.Params {
		rows = append(rows, []string{
			param.Name,
			p.inputs[i*2+3].View(),
			p.inputs[i*2+4].View(),
		})
	}

	p.paramsTable.Data(table.NewStringData(rows...))

	var title string
	switch p.mode {
	case EditCommandMode:
		title = "Edit Command"
	default:
		title = "New Command"
	}

	sty := lipgloss.NewStyle()
	return style.Border.Render(p.contentStyle.
		Width(w).
		Height(h).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				p.titleStyle.Render(title),
				p.infoTable.Render(),
				sty.MarginLeft(1).Render(style.Label.Render("Parameters")),
				sty.MarginLeft(2).Render(p.paramsTable.Render()),
			),
		))
}

// SetCommand sets the panel content.
func (p *EditPanel) SetCommand(mode EditPanelMode, cmd *command.Command) error {
	// clear the params inputs
	for _, input := range p.inputs {
		input.Reset()
	}
	p.inputs = append([]*textinput.Model{}, p.inputs[:fixedInputs]...)
	p.paramsContent = make(map[string][2]*textinput.Model)

	p.mode = mode
	if cmd == nil {
		p.cmd = command.Command{}
	} else {
		p.cmd = *cmd
		p.inputs[nameInputPos].SetValue(cmd.Name)
		p.inputs[descInputPos].SetValue(cmd.Description)
		p.inputs[cmdInputPos].SetValue(cmd.Command)
		p.refreshParamsInputs()
	}

	p.inputs[nameInputPos].Focus()
	p.selectedInput = nameInputPos
	return nil
}

// SetSize sets the panel size.
func (p *EditPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	w, _ := util.RelativeDimensions(width, height, .7, .7)
	p.infoTable.Width(w)
	p.paramsTable.Width(w)
	p.inputStyle = p.inputStyle.Width(w)
}

func (p *EditPanel) updateCommand() error {
	p.cmd.Name = p.inputs[nameInputPos].Value()
	p.cmd.Description = p.inputs[descInputPos].Value()

	cmd := p.inputs[cmdInputPos].Value()
	if len(cmd) != len(p.cmd.Command) {
		p.cmd.Command = p.inputs[cmdInputPos].Value()
		if err := p.cmd.Build(); err != nil {
			return err
		}
		p.refreshParamsInputs()
	}

	return nil
}

func (p *EditPanel) updateParams() {
	paramPos := (p.selectedInput - fixedInputs) / 2
	field := (p.selectedInput - fixedInputs) % 2

	pName := p.cmd.Params[paramPos].Name
	value := p.inputs[p.selectedInput].Value()
	p.paramsContent[pName][field].SetValue(value)

	if field == 0 {
		p.cmd.Params[paramPos].Description = value
	} else {
		p.cmd.Params[paramPos].DefaultValue = value
	}
}

func (p *EditPanel) refreshParamsInputs() {
	inputs := p.inputs[:fixedInputs]
	for _, param := range p.cmd.Params {
		if in, ok := p.paramsContent[param.Name]; ok {
			inputs = append(inputs, in[0], in[1])
		} else {
			descInput := textinput.New()
			descInput.Placeholder = "add some description"
			descInput.SetValue(param.Description)

			dvInput := textinput.New()
			dvInput.Placeholder = "optional"
			p.paramsContent[param.Name] = [2]*textinput.Model{&descInput, &dvInput}
			dvInput.SetValue(param.DefaultValue)

			inputs = append(inputs, &descInput, &dvInput)
		}
	}
	p.inputs = inputs
}
