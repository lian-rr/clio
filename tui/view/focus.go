package view

import tea "github.com/charmbracelet/bubbletea"

// Panel focus
type focus int

const (
	_ = iota
	navigationFocus
	detailFocus
	createFocus
	editFocus
	searchFocus
	executeFocus
	explainFocus
	historyFocus
)

type updateFocusMsg struct {
	UpdateFocus func(*Main)
}

func changeFocus(newFocus focus, handler func(*Main)) tea.Cmd {
	return func() tea.Msg {
		return updateFocusMsg{
			UpdateFocus: func(m *Main) {
				m.focus = newFocus
				if handler != nil {
					handler(m)
				}
			},
		}
	}
}
