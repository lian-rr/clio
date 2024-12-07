package dialog

import "github.com/charmbracelet/lipgloss"

type StyleMap struct {
	Button       lipgloss.Style
	ActiveButton lipgloss.Style
	Box          lipgloss.Style
}

var (
	defaultButton = lipgloss.NewStyle().
			Padding(0, 3).
			MarginTop(1)

	defaultStyles = StyleMap{
		Button: defaultButton.
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")),
		ActiveButton: defaultButton.
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#F25D94")).
			MarginRight(2).
			Underline(true),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true),
	}
)
