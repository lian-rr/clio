package msgs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
)

// RequestHistoryMsg is the event triggered when the command history is requested.
type RequestHistoryMsg struct {
	CommandID uuid.UUID
}

// HandleRequestHistoryMsg returns a new HandleRequestHistoryMsg.
func HandleRequestHistoryMsg(commandID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		return RequestHistoryMsg{
			CommandID: commandID,
		}
	}
}

// SetHistoryMsg returns the history content
type SetHistoryMsg struct {
	History command.History
}

// HandleRequestHistoryMsg returns a new SetHistoryMsg.
func HandleSetHistoryMsg(history command.History) tea.Cmd {
	return func() tea.Msg {
		return SetHistoryMsg{
			History: history,
		}
	}
}
