package panel

import (
	"bytes"
	"log/slog"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

// ExplainPanel handles the panel for explaing the command
type ExplainPanel struct {
	logger  *slog.Logger
	comand  string
	content viewport.Model
	spinner spinner.Model

	width   int
	height  int
	loading bool

	// styles
	titleStyle lipgloss.Style
}

func NewExplainPanel(logger *slog.Logger) ExplainPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return ExplainPanel{
		logger:     logger,
		content:    vp,
		spinner:    s,
		titleStyle: style.Title,
	}
}

func (p *ExplainPanel) Init() tea.Cmd {
	return p.spinner.Tick
}

func (p *ExplainPanel) SetCommand(cmd command.Command) error {
	var b bytes.Buffer
	if err := quick.Highlight(&b, cmd.Command, chromaLang, chromaFormatter, chromaStyle); err != nil {
		return err
	}

	p.comand = b.String()
	p.content.SetContent("")
	p.loading = true
	p.spinner.Tick()

	return nil
}

func (p *ExplainPanel) SetExplanation(explanation string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(p.width),
		glamour.WithStandardStyle(styles.DarkStyle),
	)
	if err != nil {
		return err
	}

	str, err := renderer.Render(explanation)
	if err != nil {
		return err
	}

	p.content.SetContent(str)
	p.loading = false
	return nil
}

func (p *ExplainPanel) View() string {
	sty := lipgloss.NewStyle()
	cont := "Loading " + p.spinner.View()
	if !p.loading {
		cont = lipgloss.JoinVertical(lipgloss.Center,
			sty.PaddingRight(2).
				PaddingLeft(2).
				Render(p.content.View()),
		)
	}

	return style.Border.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			p.titleStyle.Render("Explain"),
			style.Label.Render(p.comand),
			sty.PaddingTop(1).
				Render(style.Label.Render("Explanation")),
			cont,
		))
}

func (p *ExplainPanel) Update(msg tea.Msg) (ExplainPanel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		p.content, cmd = p.content.Update(msg)
	case spinner.TickMsg:
		p.spinner, cmd = p.spinner.Update(msg)
	case msgs.SetExplanationMsg:
		p.SetExplanation(msg.Explanation)
	}
	return *p, cmd
}

func (p *ExplainPanel) SetSize(width, height int) {
	p.titleStyle.Width(width)
	p.width = width
	p.height = height

	w, h := util.RelativeDimensions(width, height, .9, .77)
	p.content.Width = w
	p.content.Height = h
}
