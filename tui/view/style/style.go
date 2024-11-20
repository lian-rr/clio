package style

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor = "#F7FAF7"
	borderColor  = "#5f87ff"
)

var (
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor))

	DocStyle = BorderStyle.
			Margin(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor)).
			Foreground(lipgloss.Color(primaryColor))

	HelpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Padding(0, 2, 0).
			MarginTop(1)

	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2, 0)

	InfoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(SubtleStyle)

	HeaderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(SubtleStyle)

	LabelStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Right).
			MarginLeft(1).
			MarginRight(1).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB"))

	SubtleStyle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)
