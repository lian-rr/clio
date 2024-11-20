package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/keep/command"
)

type newCmdMsg struct {
	command command.Command
}

func handleNewCmdMsg(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return newCmdMsg{
			command: cmd,
		}
	}
}

type editCmdMsg struct {
	command command.Command
}

func handleUpdateCmd(cmd command.Command) tea.Cmd {
	return func() tea.Msg {
		return editCmdMsg{
			command: cmd,
		}
	}
}
