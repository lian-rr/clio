package panel

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
	ckey "github.com/lian-rr/clio/tui/view/key"
)

// Explorer handles the panel for listing the commands.
type Explorer struct {
	keyMap ckey.Map
	list   list.Model
}

// NewExplorer returns a new ExplorerView.
func NewExplorer(keys ckey.Map) Explorer {
	view := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	view.DisableQuitKeybindings()
	view.SetShowTitle(false)
	view.SetFilteringEnabled(false)
	view.SetShowHelp(false)
	view.SetShowStatusBar(false)

	return Explorer{
		list:   view,
		keyMap: keys,
	}
}

// Update handles the msgs.
func (p *Explorer) Update(msg tea.Msg) (Explorer, tea.Cmd) {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return *p, cmd
}

// View returns the string representation of the panel.
func (p Explorer) View() string {
	return p.list.View()
}

// SetSize sets the size the panel.
func (p *Explorer) SetSize(w, h int) {
	p.list.SetSize(w, h)
}

// SelectedCommand returns the ExplorerItem selected.
// Returns false if item not found or of incorrect type.
func (p *Explorer) SelectedCommand() (*ExplorerItem, bool) {
	command, ok := p.list.SelectedItem().(*ExplorerItem)
	if !ok {
		return nil, false
	}

	return command, true
}

// SetCommands sets the content of the list.
func (p *Explorer) SetCommands(cmds []command.Command) {
	p.list.SetItems(toListItem(cmds))
}

// AddCommand adds a new item to the List
func (p *Explorer) AddCommand(cmd command.Command) int {
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
func (p *Explorer) RemoveSelectedCommand() int {
	idx := p.list.Index()
	p.list.RemoveItem(idx)

	if idx-1 < 0 {
		return -1
	}
	return idx - 1
}

// Select selects the element in the provided index.
func (p *Explorer) Select(idx int) {
	p.list.Select(idx)
}

// RefreshCommand refresh the item command of the selected Item.
func (p *Explorer) RefreshCommand(cmd command.Command) {
	idx := p.list.Index()
	p.list.SetItem(idx, &ExplorerItem{
		title:   cmd.Name,
		desc:    cmd.Description,
		Command: &cmd,
		Loaded:  true,
	})
}

func (p *Explorer) ShortHelp() []key.Binding {
	return []key.Binding{
		p.keyMap.Quit,
		p.keyMap.Compose,
		p.keyMap.Edit,
		p.keyMap.Copy,
		p.keyMap.Delete,
		p.keyMap.Search,
		p.keyMap.DiscardSearch,
		p.keyMap.Explain,
	}
}

func (p *Explorer) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// ExplorerItem is the explorer items
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
