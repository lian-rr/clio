package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor = "#F7FAF7"
	borderColor  = "#5f87ff"
)

var (
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor))

	docStyle = borderStyle.
			Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor)).
			Foreground(lipgloss.Color(primaryColor))

	helpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Padding(0, 2, 0).
			MarginTop(1)

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2, 0)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle)

	labelStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Right).
			MarginLeft(1).
			MarginRight(1).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB"))

	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)
