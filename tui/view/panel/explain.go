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
	"github.com/lian-rr/clio/tui/view/style"
	"github.com/lian-rr/clio/tui/view/util"
)

const content = "# Summary\nThe command `git branch | grep -v \"{{.name}}\" | xargs git branch -D` is used to delete all local Git branches except for a specific branch indicated by the parameter `{{.name}}`. This command is useful for cleaning up branches in a local Git repository without having to delete each one manually.\n\n# Breakdown\n- `git branch`: This part lists all the local branches in the repository.\n- `|`: This symbol is a pipe that takes the output of the command on the left and uses it as input for the command on the right.\n- `grep -v \"{{.name}}\"`: `grep` is a command-line utility that searches for patterns in input. The `-v` option inverts the match, so it outputs all lines that do not match the specified pattern. Here, it excludes the branch named `{{.name}}` from the output.\n- `| xargs git branch -D`: The `xargs` command takes the output from the previous command (i.e., the list of branches to delete) and executes `git branch -D` on each of them. The `-D` option forces the deletion of branches without checking for unmerged changes.\n\n# Example of Use\nIf you want to delete all local branches except for `main`, you would replace `{{.name}}` with `main` in the command:\n\n```fish\ngit branch | grep -v \"main\" | xargs git branch -D\n```\n\nThis command will delete all local branches except the `main` branch.\n\n# Cautions\n- **Irreversibility**: The `-D` flag in `git branch -D` deletes branches forcefully and irreversibly. This action cannot be undone, so ensure that you really want to delete the branches.\n- **Current Branch**: Make sure you are not currently on the branch that you are trying to exclude, as this could lead to unintended behavior or errors.\n- **Branch Names**: Ensure that the branch name you are passing (i.e., `{{.name}}`) is correct and matches what you intend to keep. Any typos could lead to unwanted deletions.\n- **Safe Deletion**: If you want to review the branches before deletion, consider using `git branch | grep -v \"{{.name}}\"` alone first to see what branches will be deleted."

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

	str, err := renderer.Render(content)
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
