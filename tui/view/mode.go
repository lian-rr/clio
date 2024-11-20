package view

import tea "github.com/charmbracelet/bubbletea"

type focus int

const (
	_ = iota
	NavigationFocus
	DetailFocus
	CreateFocus
	EditFocus
	SearchFocus
	ExecuteFocus
)

type UpdateFocusMsg struct {
	UpdateFocus func(*Main)
}

func ChangeFocus(newFocus focus, handler func(*Main)) tea.Cmd {
	return func() tea.Msg {
		return UpdateFocusMsg{
			UpdateFocus: func(m *Main) {
				m.focus = newFocus
				if handler != nil {
					handler(m)
				}
			},
		}
	}
}
