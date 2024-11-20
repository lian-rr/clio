package panel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian-rr/clio/tui/view/style"
)

type SearchView struct {
	title string
	input textinput.Model
}

func NewSearchView() SearchView {
	input := textinput.New()
	input.Placeholder = "type something"
	input.TextStyle = lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#909090",
			Dark:  "#626262",
		})

	return SearchView{
		title: "Search",
		input: input,
	}
}

func (s *SearchView) Update(msg tea.Msg) (SearchView, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return *s, cmd
}

func (s *SearchView) View() string {
	return style.BorderStyle.BorderBottom(true).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			s.title+" ",
			s.input.View(),
		),
	)
}

func (s *SearchView) Focus() {
	s.input.Focus()
}

func (s *SearchView) Unfocus() {
	s.input.Blur()
}

func (s *SearchView) Reset() {
	s.input.Reset()
}

func (s *SearchView) Content() string {
	return s.input.Value()
}

func (s *SearchView) SetWidth(width int) {
	s.input.Width = width - (len(s.title) + 4)
}
