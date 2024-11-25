package msgs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

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
	CommandID   uuid.UUID
	Explanation string
	Cache       bool
}

// HandleNewCommandMsg returns a new SetExplanationMsg.
func HandleSetExplanationMsg(commandID uuid.UUID, explanation string, cache bool) tea.Cmd {
	return func() tea.Msg {
		return SetExplanationMsg{
			CommandID:   commandID,
			Explanation: explanation,
			Cache:       cache,
		}
	}
}

// CacheExplanationMsg is the event triggered for caching the command explanation
type CacheExplanationMsg struct {
	CommandID   uuid.UUID
	Explanation string
}

// HandleNewCommandMsg returns a new CacheExplanationMsg.
func HandleCacheExplanationMsg(commandID uuid.UUID, explanation string) tea.Cmd {
	return func() tea.Msg {
		return CacheExplanationMsg{
			CommandID:   commandID,
			Explanation: explanation,
		}
	}
}

// EvictCachedExplanationMsg is the event triggered for deleting the cached command explanation
type EvictCachedExplanationMsg struct {
	CommandID uuid.UUID
}

// HandleEvictCachedExplanationMsg returns a new EvictCachedExplanationMsg.
func HandleEvictCachedExplanationMsg(commandID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		return EvictCachedExplanationMsg{
			CommandID: commandID,
		}
	}
}
