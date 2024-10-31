package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchView struct {
	title string
	input textinput.Model
}

func newSearchView() searchView {
	input := textinput.New()
	input.Placeholder = "type / to search"
	input.TextStyle = lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#909090",
			Dark:  "#626262",
		})

	return searchView{
		title: "Search",
		input: input,
	}
}

func (s *searchView) Update(msg tea.Msg) (searchView, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return *s, cmd
}

func (s *searchView) View() string {
	return borderStyle.BorderBottom(true).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			s.title+" ",
			s.input.View(),
		),
	)
}

func (s *searchView) Focus() {
	s.input.Focus()
}

func (s *searchView) Content() string {
	return s.input.Value()
}

func (s *searchView) SetWidth(width int) {
	s.input.Width = width - (len(s.title) + 4)
}
