package panel

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
)

type ListView struct {
	list list.Model
}

func NewListView() ListView {
	view := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	view.DisableQuitKeybindings()
	view.SetShowTitle(false)
	view.SetFilteringEnabled(false)
	view.SetShowHelp(false)
	view.SetShowStatusBar(false)

	return ListView{
		list: view,
	}
}

func (l *ListView) Update(msg tea.Msg) (ListView, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

func (l ListView) View() string {
	return l.list.View()
}

func (l *ListView) SetSize(w, h int) {
	l.list.SetSize(w, h)
}

func (l *ListView) SelectedItem() (*ListItem, bool) {
	command, ok := l.list.SelectedItem().(*ListItem)
	if !ok {
		return nil, false
	}

	return command, true
}

func (l *ListView) SetContent(cmds []command.Command) {
	l.list.SetItems(toListItem(cmds))
}

func (l *ListView) AddItem(cmd command.Command) int {
	idx := len(l.list.Items())
	l.list.InsertItem(idx, &ListItem{
		title:  cmd.Name,
		desc:   cmd.Description,
		Cmd:    &cmd,
		Loaded: true,
	})

	return idx
}

func (l *ListView) RemoveSelectedItem() int {
	idx := l.list.Index()
	l.list.RemoveItem(idx)

	if idx-1 < 0 {
		return -1
	}
	return idx - 1
}

func (l *ListView) Select(idx int) {
	l.list.Select(idx)
}

func (l *ListView) RefreshItem(cmd command.Command) {
	idx := l.list.Index()
	l.list.SetItem(idx, ListItem{
		title: cmd.Name,
		desc:  cmd.Description,
		Cmd:   &cmd,
	})
}

func toListItem(cmds []command.Command) []list.Item {
	items := make([]list.Item, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, &ListItem{
			title: cmd.Name,
			desc:  cmd.Description,
			Cmd:   &cmd,
		})
	}

	return items
}

type ListItem struct {
	title  string
	desc   string
	Cmd    *command.Command
	Loaded bool
}

var _ list.Item = (*ListItem)(nil)

func (i ListItem) Title() string {
	return i.title
}

func (i ListItem) Description() string {
	return i.desc
}

func (i ListItem) FilterValue() string {
	return i.title
}
