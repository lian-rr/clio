package panel

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/view/event"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

// ExecutePanel handles the panel for executing a command.
type ExecutePanel struct {
	command *command.Command

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

// NewExecutePanel returns a new ExecutePanel.
func NewExecutePanel(logger *slog.Logger) ExecutePanel {
	infoTable := table.New().
		Border(lipgloss.HiddenBorder())

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("NAME", "DESCRIPTION", "DEFAULT VALUE")

	return ExecutePanel{
		logger:      logger,
		infoTable:   infoTable,
		paramsTable: params,
		titleStyle: style.LabelStyle.
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(style.SubtleStyle),
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

// Update handles the msgs.
func (v *ExecutePanel) Update(msg tea.KeyMsg) (ExecutePanel, tea.Cmd) {
	paramCount := len(v.paramInputs)
	var cmd tea.Cmd
	if paramCount != 0 {
		var input textinput.Model
		switch {
		case key.Matches(msg, ckey.DefaultMap.NextParamKey):
			v.paramInputs[v.orderedParams[v.selectedInput]].Blur()
			v.selectedInput = (v.selectedInput + 1) % paramCount
			v.paramInputs[v.orderedParams[v.selectedInput]].Focus()
		case key.Matches(msg, ckey.DefaultMap.PreviousParamKey):
			v.paramInputs[v.orderedParams[v.selectedInput]].Blur()
			// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
			v.selectedInput = ((v.selectedInput-1)%paramCount + paramCount) % paramCount
			v.paramInputs[v.orderedParams[v.selectedInput]].Focus()
		case key.Matches(msg, ckey.DefaultMap.Enter):
			v.paramInputs[v.orderedParams[v.selectedInput]].Blur()
			out, err := v.produceCommand()
			if err != nil {
				v.logger.Warn("producing incomplete command", slog.Any("error", err))
				break
			}
			return *v, event.HandleExecuteMsg(out)
		default:
			param := v.orderedParams[v.selectedInput]
			input, cmd = v.paramInputs[param].Update(msg)
			v.paramInputs[param] = &input
		}
	}
	return *v, cmd
}

// View returns the string representation of the panel.
func (v *ExecutePanel) View() string {
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

	return style.BorderStyle.Render(v.contentStyle.
		Width(w).
		Height(h).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				v.titleStyle.Render("Compose"),
				v.infoTable.Render(),
				style.BorderStyle.Render(outCommand),
				v.paramsTable.Render(),
			),
		))
}

// SetCommand sets the panel content.
func (v *ExecutePanel) SetCommand(cmd command.Command) error {
	inputStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#2aa198",
			Dark:  "#2aa198",
		})

	v.command = &cmd

	v.infoTable.Data(table.NewStringData([][]string{
		{style.LabelStyle.Render("Name"), cmd.Name},
		{style.LabelStyle.Render("Description"), cmd.Description},
	}...))

	rows := make([][]string, 0, len(cmd.Params))
	orderedParams := make([]string, 0, len(cmd.Params))
	v.paramInputs = make(map[string]*textinput.Model, len(cmd.Params))
	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})

		pi := textinput.New()
		pi.Placeholder = param.Name
		pi.TextStyle = inputStyle
		pi.Prompt = ""
		pi.CharLimit = 32
		if param.DefaultValue != "" {
			pi.SetValue(param.DefaultValue)
		}

		v.paramInputs[param.Name] = &pi
		orderedParams = append(orderedParams, param.Name)
	}

	v.paramsTable.Data(table.NewStringData(rows...))
	v.orderedParams = orderedParams
	if len(orderedParams) > 0 {
		v.paramInputs[orderedParams[0]].Focus()
	}

	return nil
}

// SetSize sets the panel size.
func (v *ExecutePanel) SetSize(width, height int) {
	v.width = width
	v.height = height
	w, _ := util.RelativeDimensions(width, height, .7, .7)
	v.infoTable.Width(w)
	v.paramsTable.Width(w)
}

func (v *ExecutePanel) produceCommand() (string, error) {
	arguments := make([]command.Argument, 0, len(v.command.Params))
	for param, input := range v.paramInputs {
		val := input.Value()
		if len(val) == 0 {
			return "", fmt.Errorf("value empty for param %q", param)
		}
		arguments = append(arguments, command.Argument{
			Name:  param,
			Value: val,
		})
	}

	outCommand, err := v.command.Compile(arguments)
	if err != nil {
		return "", fmt.Errorf("error compiling command: %v", err)
	}

	return outCommand, nil
}
