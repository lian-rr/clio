package panel

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
)

// ExplorerPanel handles the panel for listing the commands.
type ExplorerPanel struct {
	list list.Model
}

// NewExplorerPanel returns a new ExplorerView.
func NewExplorerPanel() ExplorerPanel {
	view := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	view.DisableQuitKeybindings()
	view.SetShowTitle(false)
	view.SetFilteringEnabled(false)
	view.SetShowHelp(false)
	view.SetShowStatusBar(false)

	return ExplorerPanel{
		list: view,
	}
}

// Update handles the msgs.
func (l *ExplorerPanel) Update(msg tea.Msg) (ExplorerPanel, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

// View returns the string representation of the panel.
func (l ExplorerPanel) View() string {
	return l.list.View()
}

// SetSize sets the size the panel.
func (l *ExplorerPanel) SetSize(w, h int) {
	l.list.SetSize(w, h)
}

// SelectedCommand returns the ExplorerItem selected.
// Returns false if item not found or of incorrect type.
func (l *ExplorerPanel) SelectedCommand() (*ExplorerItem, bool) {
	command, ok := l.list.SelectedItem().(*ExplorerItem)
	if !ok {
		return nil, false
	}

	return command, true
}

// SetCommands sets the content of the list.
func (l *ExplorerPanel) SetCommands(cmds []command.Command) {
	l.list.SetItems(toListItem(cmds))
}

// AddCommand adds a new item to the List
func (l *ExplorerPanel) AddCommand(cmd command.Command) int {
	idx := len(l.list.Items())
	l.list.InsertItem(idx, &ExplorerItem{
		title:   cmd.Name,
		desc:    cmd.Description,
		Command: &cmd,
		Loaded:  true,
	})

	return idx
}

// RemoveSelectedCommand removes the selected item form the list.
func (l *ExplorerPanel) RemoveSelectedCommand() int {
	idx := l.list.Index()
	l.list.RemoveItem(idx)

	if idx-1 < 0 {
		return -1
	}
	return idx - 1
}

// Select selects the element in the provided index.
func (l *ExplorerPanel) Select(idx int) {
	l.list.Select(idx)
}

// RefreshCommand refresh the item command of the selected Item.
func (l *ExplorerPanel) RefreshCommand(cmd command.Command) {
	idx := l.list.Index()
	l.list.SetItem(idx, ExplorerItem{
		title:   cmd.Name,
		desc:    cmd.Description,
		Command: &cmd,
		Loaded:  true,
	})
}

func toListItem(cmds []command.Command) []list.Item {
	items := make([]list.Item, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, &ExplorerItem{
			title:   cmd.Name,
			desc:    cmd.Description,
			Command: &cmd,
		})
	}

	return items
}

type ExplorerItem struct {
	title string
	desc  string

	// Command is a pointer to the represented command.
	Command *command.Command
	// Loaded indicates if the inner command has been fully loaded.
	Loaded bool
}

var _ list.Item = (*ExplorerItem)(nil)

func (i ExplorerItem) Title() string {
	return i.title
}

func (i ExplorerItem) Description() string {
	return i.desc
}

func (i ExplorerItem) FilterValue() string {
	return i.title
}
