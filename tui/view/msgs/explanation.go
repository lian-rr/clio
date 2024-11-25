package msgs

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
)

// RequestExplanationMsg is the event triggered when the command explanation is requested.
type RequestExplanationMsg struct {
	Command command.Command
}

// HandleRequestExplanationMsg returns a new RequestExplanationMsg.
func HandleRequestExplanationMsg(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return RequestExplanationMsg{
			Command: cmd,
		}
	}
}

// SetExplanationMsg is the event triggered for setting the command explanation
type SetExplanationMsg struct {
	Explanation string
}

// HandleNewCommandMsg returns a new SetExplanationMsg.
func HandleSetExplanationMsg(explanation string) tea.Cmd {
	return func() tea.Msg {
		return SetExplanationMsg{
			Explanation: explanation,
		}
	}
}
