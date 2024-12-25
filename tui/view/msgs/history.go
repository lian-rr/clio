package msgs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
)

// SaveUsageMsg is the event triggered for saving the usage of the command.
type SaveUsageMsg struct {
	CommandID uuid.UUID
	Usage     string
}

// HandleSaveUsageMsg returns a new SaveUsageMsg
func HandleSaveUsageMsg(commandID uuid.UUID, usage string) tea.Cmd {
	return func() tea.Msg {
		return SaveUsageMsg{
			CommandID: commandID,
			Usage:     usage,
		}
	}
}

// RequestHistoryMsg is the event triggered when the command history is requested.
type RequestHistoryMsg struct {
	CommandID uuid.UUID
}

// HandleRequestHistoryMsg returns a new RequestHistoryMsg.
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
