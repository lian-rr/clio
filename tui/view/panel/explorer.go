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
func (p *ExplorerPanel) Update(msg tea.Msg) (ExplorerPanel, tea.Cmd) {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return *p, cmd
}

// View returns the string representation of the panel.
func (p ExplorerPanel) View() string {
	return p.list.View()
}

// SetSize sets the size the panel.
func (p *ExplorerPanel) SetSize(w, h int) {
	p.list.SetSize(w, h)
}

// SelectedCommand returns the ExplorerItem selected.
// Returns false if item not found or of incorrect type.
func (p *ExplorerPanel) SelectedCommand() (*ExplorerItem, bool) {
	command, ok := p.list.SelectedItem().(*ExplorerItem)
	if !ok {
		return nil, false
	}

	return command, true
}

// SetCommands sets the content of the list.
func (p *ExplorerPanel) SetCommands(cmds []command.Command) {
	p.list.SetItems(toListItem(cmds))
}

// AddCommand adds a new item to the List
func (p *ExplorerPanel) AddCommand(cmd command.Command) int {
	idx := len(p.list.Items())
	p.list.InsertItem(idx, &ExplorerItem{
		title:   cmd.Name,
		desc:    cmd.Description,
		Command: &cmd,
		Loaded:  true,
	})

	return idx
}

// RemoveSelectedCommand removes the selected item form the list.
func (p *ExplorerPanel) RemoveSelectedCommand() int {
	idx := p.list.Index()
	p.list.RemoveItem(idx)

	if idx-1 < 0 {
		return -1
	}
	return idx - 1
}

// Select selects the element in the provided index.
func (p *ExplorerPanel) Select(idx int) {
	p.list.Select(idx)
}

// RefreshCommand refresh the item command of the selected Item.
func (p *ExplorerPanel) RefreshCommand(cmd command.Command) {
	idx := p.list.Index()
	p.list.SetItem(idx, &ExplorerItem{
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
