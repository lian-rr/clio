package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian_rr/keep/command"
)

type listView struct {
	list list.Model
}

func newListView() listView {
	view := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	view.DisableQuitKeybindings()
	view.SetShowTitle(false)
	view.SetFilteringEnabled(false)
	view.SetShowHelp(false)
	view.SetShowStatusBar(false)

	return listView{
		list: view,
	}
}

func (l *listView) Update(msg tea.Msg) (listView, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

func (l listView) View() string {
	return l.list.View()
}

func (l *listView) SetSize(w, h int) {
	l.list.SetSize(w, h)
}

func (l *listView) selectedItem() (*listItem, bool) {
	command, ok := l.list.SelectedItem().(*listItem)
	if !ok {
		return nil, false
	}

	return command, true
}

func (l *listView) SetContent(cmds []command.Command) {
	l.list.SetItems(toListItem(cmds))
}

func toListItem(cmds []command.Command) []list.Item {
	items := make([]list.Item, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, &listItem{
			title: cmd.Name,
			desc:  cmd.Description,
			cmd:   &cmd,
		})
	}

	return items
}

type listItem struct {
	title  string
	desc   string
	cmd    *command.Command
	loaded bool
}

var _ list.Item = (*listItem)(nil)

func (i listItem) Title() string {
	return i.title
}

func (i listItem) Description() string {
	return i.desc
}

func (i listItem) FilterValue() string {
	return i.title
}
