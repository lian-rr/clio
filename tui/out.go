package tui

import tea "github.com/charmbracelet/bubbletea"

type outcomeMsg struct {
	outcome string
}

func handleOutcome(out string) tea.Cmd {
	return func() tea.Msg {
		return outcomeMsg{
			outcome: out,
		}
	}
}
