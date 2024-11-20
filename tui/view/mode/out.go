package mode

import tea "github.com/charmbracelet/bubbletea"

type OutcomeMsg struct {
	Output string
}

func HandleOutcome(out string) tea.Cmd {
	return func() tea.Msg {
		return OutcomeMsg{
			Output: out,
		}
	}
}
