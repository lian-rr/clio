package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type executeView struct {
	paramsTable *table.Table
}

func (v *executeView) Init() tea.Cmd {
	return nil
}

func (v *executeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (v *executeView) View() string {
	return lipgloss.JoinVertical(lipgloss.Center, "hello there!", "don't talk to me")
}
