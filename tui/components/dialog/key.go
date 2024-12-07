package dialog

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Discard  key.Binding
	Accept   key.Binding
	Navigate key.Binding
}

var defaultKeyMap = KeyMap{
	Discard: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "discard"),
	),
	Accept: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "accept"),
	),
	Navigate: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "navigate"),
	),
}
