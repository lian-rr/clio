package panel

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

// Execute handles the panel for executing a command.
type Execute struct {
	command *command.Command
	keyMap  ckey.Map

	paramsTable *table.Table
	infoTable   *table.Table
	paramInputs map[string]*textinput.Model

	orderedParams []string
	selectedInput int
	width         int
	height        int

	contentStyle lipgloss.Style
	titleStyle   lipgloss.Style
	logger       *slog.Logger
}

// NewExecute returns a new ExecutePanel.
func NewExecute(keys ckey.Map, logger *slog.Logger) Execute {
	infoTable := table.New().
		Border(lipgloss.HiddenBorder())

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("NAME", "DESCRIPTION", "DEFAULT VALUE")

	return Execute{
		keyMap:      keys,
		logger:      logger,
		infoTable:   infoTable,
		paramsTable: params,
		titleStyle: style.Label.
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(style.Subtle),
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

// Init starts the input blink
func (p *Execute) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles the msgs.
func (p *Execute) Update(msg tea.Msg) (Execute, tea.Cmd) {
	paramCount := len(p.paramInputs)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keyMap.NextParamKey):
			if paramCount > 1 {
				p.paramInputs[p.orderedParams[p.selectedInput]].Blur()
				p.selectedInput = (p.selectedInput + 1) % paramCount
				p.paramInputs[p.orderedParams[p.selectedInput]].Focus()
			}
		case key.Matches(msg, p.keyMap.PreviousParamKey):
			if paramCount > 1 {
				p.paramInputs[p.orderedParams[p.selectedInput]].Blur()
				// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
				p.selectedInput = ((p.selectedInput-1)%paramCount + paramCount) % paramCount
				p.paramInputs[p.orderedParams[p.selectedInput]].Focus()
			}
		case key.Matches(msg, p.keyMap.Go):
			id, out, err := p.produceCommand()
			if err != nil {
				p.logger.Warn("producing incomplete command", slog.Any("error", err))
				break
			}
			return *p, msgs.HandleExecuteMsg(id, out)
		default:
			if len(p.paramInputs) > 0 {
				var input textinput.Model
				param := p.orderedParams[p.selectedInput]
				input, cmd = p.paramInputs[param].Update(msg)
				p.paramInputs[param] = &input
			}
		}
	default:
		// handling blinking mostly
		if len(p.paramInputs) > 0 {
			var input textinput.Model
			param := p.orderedParams[p.selectedInput]
			input, cmd = p.paramInputs[param].Update(msg)
			p.paramInputs[param] = &input
		}
	}
	return *p, cmd
}

// View returns the string representation of the panel.
func (p *Execute) View() string {
	if p.command == nil {
		return ""
	}

	arguments := make([]command.Argument, 0, len(p.command.Params))
	for param, input := range p.paramInputs {
		arguments = append(arguments, command.Argument{
			Name:  param,
			Value: input.View(),
		})
	}

	outCommand, err := p.command.Compile(arguments)
	if err != nil {
		p.logger.Error("error compiling command",
			slog.String("name", p.command.Name),
			slog.String("command", p.command.Command),
			slog.Any("params", p.command.Params),
			slog.Any("arguments", arguments),
			slog.Any("error", err),
		)
		return ""
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
					p.titleStyle.Render("Compose"),
					p.infoTable.Render(),
					style.Border.Render(outCommand),
					p.paramsTable.Render(),
				),
			))
}

// SetCommand sets the panel content.
func (p *Execute) SetCommand(cmd command.Command) error {
	inputStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#2aa198",
			Dark:  "#2aa198",
		})

	p.command = &cmd

	p.infoTable.Data(table.NewStringData([][]string{
		{style.Label.Render("Name"), cmd.Name},
		{style.Label.Render("Description"), cmd.Description},
	}...))

	rows := make([][]string, 0, len(cmd.Params))
	orderedParams := make([]string, 0, len(cmd.Params))
	p.paramInputs = make(map[string]*textinput.Model, len(cmd.Params))

	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})

		pi := textinput.New()
		pi.Placeholder = param.Name
		pi.TextStyle = inputStyle
		pi.Prompt = ""
		pi.CharLimit = 32
		if param.DefaultValue != "" {
			pi.SetValue(param.DefaultValue)
			pi.SetCursor(len(param.DefaultValue))
		}

		p.paramInputs[param.Name] = &pi
		orderedParams = append(orderedParams, param.Name)
	}

	p.paramsTable.Data(table.NewStringData(rows...))
	p.orderedParams = orderedParams
	if len(orderedParams) > 0 {
		p.paramInputs[orderedParams[0]].Focus()
	}

	p.logger.Debug("command to execute set", slog.Any("command", cmd))
	return nil
}

// SetSize sets the panel size.
func (p *Execute) SetSize(width, height int) {
	p.width = width
	p.height = height
	w, _ := util.RelativeDimensions(width, height, .7, .7)
	p.paramsTable.Width(w)
}

func (p *Edit) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Back,
		p.keyMap.NextParamKey,
		p.keyMap.PreviousParamKey,
		p.keyMap.Go,
	}
}

func (p *Edit) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func (p *Execute) produceCommand() (commandID uuid.UUID, out string, err error) {
	arguments := make([]command.Argument, 0, len(p.command.Params))
	for param, input := range p.paramInputs {
		val := input.Value()
		if len(val) == 0 {
			return uuid.Nil, "", fmt.Errorf("value empty for param %q", param)
		}
		arguments = append(arguments, command.Argument{
			Name:  param,
			Value: val,
		})
	}

	outCommand, err := p.command.Compile(arguments)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("error compiling command: %p", err)
	}

	return p.command.ID, outCommand, nil
}
