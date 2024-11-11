package tui

import tea "github.com/charmbracelet/bubbletea"

type mode int

const (
	_ = iota
	navigationMode
	detailMode
	createMode
	editMode
	searchMode
	executeMode
)

type updateModeMsg struct {
	updateMode func(*view)
}

func changeMode(newMode mode, handler func(*view)) tea.Cmd {
	return func() tea.Msg {
		return updateModeMsg{
			updateMode: func(m *view) {
				m.currentMode = newMode
				if handler != nil {
					handler(m)
				}
			},
		}
	}
}
