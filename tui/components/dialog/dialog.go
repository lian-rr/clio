package dialog

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dialog confirmation component model.
type Dialog struct {
	text              string
	acceptButtonLabel string
	cancelButtonLabel string

	width       int
	height      int
	selectedBtn int

	keys  KeyMap
	style StyleMap
}

// New returns a new Dialog
func New(text string, opts ...OptFunc) Dialog {
	dialog := Dialog{
		text:              text,
		acceptButtonLabel: "Accept",
		cancelButtonLabel: "Cancel",
		keys:              defaultKeyMap,
		style:             defaultStyles,
	}

	for _, opt := range opts {
		opt(&dialog)
	}

	return dialog
}

// Update handles the input events.
func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Navigate):
			d.selectedBtn = (d.selectedBtn + 1) % 2
		case key.Matches(msg, d.keys.Accept):
			if d.selectedBtn == 0 {
				cmd = sendMsg(AcceptMsg{})
			} else {
				cmd = sendMsg(DiscardMsg{})
			}
		case key.Matches(msg, d.keys.Discard):
			cmd = sendMsg(DiscardMsg{})
		}
	}
	return d, cmd
}

// View renders the Dialog content.
func (d Dialog) View() string {
	var acceptBtn, cancelBtn string
	if d.selectedBtn == 0 {
		acceptBtn = d.style.ActiveButton.Render(d.acceptButtonLabel)
		cancelBtn = d.style.Button.Render(d.cancelButtonLabel)
	} else {
		acceptBtn = d.style.Button.Render(d.acceptButtonLabel)
		cancelBtn = d.style.ActiveButton.Render(d.cancelButtonLabel)
	}

	question := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(d.text)

	ui := lipgloss.JoinVertical(
		lipgloss.Center,
		question,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			acceptBtn,
			cancelBtn,
		),
	)

	content := lipgloss.NewStyle().
		Width(50).
		AlignHorizontal(lipgloss.Center).
		Render(d.style.Box.Render(ui))

	return content
}

// Init
func (d Dialog) Init() tea.Cmd {
	return sendMsg(InitMsg{})
}

// Reset resets the Dialog options.
func (d Dialog) Reset() Dialog {
	d.selectedBtn = 0
	return d
}

type (
	// InitMsg send when the confirmation panenl is opened
	InitMsg struct{}
	// AcceptMsg send when the accept event triggerd
	AcceptMsg struct{}
	// DiscardMsg send when the discard event triggerd
	DiscardMsg struct{}
)

func sendMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
