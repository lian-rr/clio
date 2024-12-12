package key

import (
	"github.com/charmbracelet/bubbles/key"
)

// Map of key bindings.
type Map struct {
	Search           key.Binding
	DiscardSearch    key.Binding
	Quit             key.Binding
	ForceQuit        key.Binding
	Compose          key.Binding
	Go               key.Binding
	Back             key.Binding
	New              key.Binding
	Edit             key.Binding
	Explain          key.Binding
	Copy             key.Binding
	NextParamKey     key.Binding
	PreviousParamKey key.Binding
	Delete           key.Binding
}

// DefaultMap of key bindings.
var DefaultMap = Map{
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search")),
	DiscardSearch: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "discard search")),
	Quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "quit")),
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "force exit")),
	Compose: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "compose")),
	Go: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "go")),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back")),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new")),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit")),
	Explain: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "explain")),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy")),
	Delete: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete")),
	NextParamKey: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	PreviousParamKey: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	),
}

func (km Map) ShortHelp() []key.Binding {
	return []key.Binding{
		km.Back,
		km.Search,
		km.DiscardSearch,
		km.Compose,
		km.New,
		km.Edit,
		km.Copy,
		km.Explain,
		km.Delete,
	}
}

func (km Map) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
