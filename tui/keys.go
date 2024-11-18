package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	search           key.Binding
	discardSearch    key.Binding
	quit             key.Binding
	forceQuit        key.Binding
	enter            key.Binding
	back             key.Binding
	new              key.Binding
	nextParamKey     key.Binding
	previousParamKey key.Binding
	delete           key.Binding
}

var defaultKeyMap = keyMap{
	search: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "search")),
	discardSearch: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "all cmds")),
	quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "quit")),
	forceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "force exit")),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "compose")),
	back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back")),
	new: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new")),
	delete: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete")),
	nextParamKey: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next param"),
	),
	previousParamKey: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous param"),
	),
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.back,
		km.search,
		km.discardSearch,
		km.enter,
		km.new,
		km.delete,
	}
}

func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
