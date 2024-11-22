package msgs

import tea "github.com/charmbracelet/bubbletea"

type AsyncMsg struct {
	Msg tea.Msg
}

func AsyncHandler(activityChan <-chan AsyncMsg) tea.Cmd {
	return func() tea.Msg {
		return AsyncMsg(<-activityChan)
	}
}

func PublishAsyncMsg(activityChan chan AsyncMsg, cmd tea.Cmd) {
	if cmd != nil {
		activityChan <- AsyncMsg{
			Msg: cmd(),
		}
	}
}
