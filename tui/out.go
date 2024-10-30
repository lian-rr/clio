package tui

import tea "github.com/charmbracelet/bubbletea"

type outcomeMsg struct {
	output string
}

func handleOutcome(out string) tea.Cmd {
	return func() tea.Msg {
		return outcomeMsg{
			output: out,
		}
	}
}
