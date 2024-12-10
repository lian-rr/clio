package panel

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian-rr/clio/tui/view/style"
)

type Search struct {
	logger *slog.Logger
	title  string
	input  textinput.Model
}

func NewSearch(logger *slog.Logger) Search {
	input := textinput.New()
	input.Placeholder = "type something"
	input.TextStyle = lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#909090",
			Dark:  "#626262",
		})

	return Search{
		title:  "Search",
		input:  input,
		logger: logger,
	}
}

// Init starts the input blink
func (p *Search) Init() tea.Cmd {
	return textinput.Blink
}

func (p *Search) Update(msg tea.Msg) (Search, tea.Cmd) {
	p.logger.Debug("update in search", slog.Any("msg", msg))
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return *p, cmd
}

func (p *Search) View() string {
	return style.Border.BorderBottom(true).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			p.title+" ",
			p.input.View(),
		),
	)
}

func (p *Search) Focus() {
	p.input.Focus()
}

func (p *Search) Unfocus() {
	p.input.Blur()
}

func (p *Search) Reset() {
	p.input.Reset()
}

func (p *Search) Content() string {
	return p.input.Value()
}

func (p *Search) SetWidth(width int) {
	p.input.Width = width - (len(p.title) + 4)
}
