package ckey

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Search           key.Binding
	DiscardSearch    key.Binding
	Quit             key.Binding
	ForceQuit        key.Binding
	Enter            key.Binding
	Back             key.Binding
	New              key.Binding
	Edit             key.Binding
	NextParamKey     key.Binding
	PreviousParamKey key.Binding
	Delete           key.Binding
}

var DefaultKeyMap = KeyMap{
	Search: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "search")),
	DiscardSearch: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "all cmds")),
	Quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "quit")),
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "force exit")),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "compose")),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back")),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new")),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit")),
	Delete: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete")),
	NextParamKey: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next param"),
	),
	PreviousParamKey: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous param"),
	),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.Back,
		km.Search,
		km.DiscardSearch,
		km.Enter,
		km.New,
		km.Edit,
		km.Delete,
	}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
