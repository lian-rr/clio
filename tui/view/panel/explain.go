package panel

import (
	"bytes"
	"log/slog"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian-rr/clio/command"
	ckey "github.com/lian-rr/clio/tui/view/key"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

// Explain handles the panel for explaing the command
type Explain struct {
	logger  *slog.Logger
	keyMap  ckey.Map
	comand  string
	content viewport.Model
	spinner spinner.Model

	width   int
	height  int
	loading bool

	// styles
	titleStyle lipgloss.Style
}

func NewExplain(keys ckey.Map, logger *slog.Logger) Explain {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Explain{
		logger:     logger,
		keyMap:     keys,
		content:    vp,
		spinner:    s,
		titleStyle: style.Title,
	}
}

func (p *Explain) Init() tea.Cmd {
	return p.spinner.Tick
}

func (p *Explain) SetCommand(cmd command.Command) error {
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

func (p *Explain) SetExplanation(explanation string) error {
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

func (p *Explain) View() string {
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

func (p *Explain) Update(msg tea.Msg) (Explain, tea.Cmd) {
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

func (p *Explain) SetSize(width, height int) {
	p.titleStyle.Width(width)
	p.width = width
	p.height = height

	w, h := util.RelativeDimensions(width, height, .9, .77)
	p.content.Width = w
	p.content.Height = h
}

func (p *Explain) ShortHelp() []key.Binding {
	keys := []key.Binding{
		p.keyMap.Back,
		p.content.KeyMap.Down,
		p.content.KeyMap.Up,
		p.content.KeyMap.PageDown,
		p.content.KeyMap.PageUp,
		p.content.KeyMap.HalfPageDown,
		p.content.KeyMap.HalfPageUp,
	}
	return keys
}

func (p *Explain) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
