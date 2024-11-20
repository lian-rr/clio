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

func (p *SearchView) Update(msg tea.Msg) (SearchView, tea.Cmd) {
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return *p, cmd
}

func (p *SearchView) View() string {
	return style.BorderStyle.BorderBottom(true).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			p.title+" ",
			p.input.View(),
		),
	)
}

func (p *SearchView) Focus() {
	p.input.Focus()
}

func (p *SearchView) Unfocus() {
	p.input.Blur()
}

func (p *SearchView) Reset() {
	p.input.Reset()
}

func (p *SearchView) Content() string {
	return p.input.Value()
}

func (p *SearchView) SetWidth(width int) {
	p.input.Width = width - (len(p.title) + 4)
}
