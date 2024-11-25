package msgs

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
)

// NewCommandMsg is the event triggered for creating a new command.
type NewCommandMsg struct {
	Command command.Command
}

// HandleNewCommandMsg returns a new NewCommandMsg.
func HandleNewCommandMsg(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return NewCommandMsg{
			Command: cmd,
		}
	}
}

// UpdateCommandMsg is the event triggered for editing a command.
type UpdateCommandMsg struct {
	Command command.Command
}

// HandleUpdateCommandMsg returns a new UpdateCommandMsg.
func HandleUpdateCommandMsg(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return UpdateCommandMsg{
			Command: cmd,
		}
	}
}

// ExecuteCommandMsg is the event triggered for executing a command
type ExecuteCommandMsg struct {
	Command string
}

// HandleExecuteMsg returns a new ExecuteCommandMsg
func HandleExecuteMsg(cmd string) tea.Cmd {
	return func() tea.Msg {
		return ExecuteCommandMsg{
			Command: cmd,
		}
	}
}
