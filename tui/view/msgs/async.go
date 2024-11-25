package msgs

import tea "github.com/charmbracelet/bubbletea"

// AsyncMsg wraps the tea.Msg that are going to be handle asynchronously.
type AsyncMsg struct {
	Msg tea.Msg
}

// AsyncHandler returns a AsyncMsg from the activityChan.
func AsyncHandler(activityChan <-chan AsyncMsg) tea.Cmd {
	return func() tea.Msg {
		return AsyncMsg(<-activityChan)
	}
}

// PublishAsyncMsg sends a tea.Msg through the activityChan.
func PublishAsyncMsg(activityChan chan AsyncMsg, cmd tea.Cmd) {
	if cmd != nil {
		activityChan <- AsyncMsg{
			Msg: cmd(),
		}
	}
}
