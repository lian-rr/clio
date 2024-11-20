package mode

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
)

type NewCmdMsg struct {
	Command command.Command
}

func HandleNewCmdMsg(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return NewCmdMsg{
			Command: cmd,
		}
	}
}

type EditCmdMsg struct {
	Command command.Command
}

func HandleUpdateCmd(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return EditCmdMsg{
			Command: cmd,
		}
	}
}
