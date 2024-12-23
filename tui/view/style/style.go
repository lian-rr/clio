package style

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor = "#F7FAF7"
	borderColor  = "#5f87ff"
)

var (
	Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	Document = Border.
			Margin(1, 2)

	Title = lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Foreground(lipgloss.Color(primaryColor))

	Help = lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Padding(0, 2, 0).
		MarginTop(1)

	Container = lipgloss.NewStyle().
			Padding(1, 2, 0)

	Info = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(Subtle)

	Header = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Subtle)

	Label = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Right).
		MarginLeft(1).
		MarginRight(1).
		Padding(0, 1).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB"))

	Subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)
