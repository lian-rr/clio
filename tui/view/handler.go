package view

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/tui/view/panel"
)

func (m *Main) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch m.focus {
	case SearchFocus:
		return m.handleSearchInput(msg)
	case ExecuteFocus:
		return m.handleExecuteInput(msg)
	case EditFocus:
		return m.handleEditInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}
}

func (m *Main) handleNavigationInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Quit):
		return tea.Quit
	case key.Matches(msg, m.keys.Search):
		return ChangeFocus(SearchFocus, func(m *Main) {
			m.searchView.Focus()
		})
	case key.Matches(msg, m.keys.DiscardSearch):
		m.searchView.Reset()
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands", slog.Any("error", err))
			break
		}

		m.setContent(cmds)
	case key.Matches(msg, m.keys.Enter):
		item, ok := m.commandsView.SelectedItem()
		if !ok {
			break
		}

		return ChangeFocus(ExecuteFocus, func(m *Main) {
			err := m.executeView.SetCommand(*item.Cmd)
			if err != nil {
				m.logger.Error("error setting execute view content", slog.Any("error", err))
			}
		})
	case key.Matches(msg, m.keys.New):
		return ChangeFocus(EditFocus, func(m *Main) {
			err := m.editView.SetCommand(panel.NewCommandMode, nil)
			if err != nil {
				m.logger.Error("error setting edit view content", slog.Any("error", err))
			}
		})
	case key.Matches(msg, m.keys.Edit):
		item, ok := m.commandsView.SelectedItem()
		if !ok {
			break
		}

		return ChangeFocus(EditFocus, func(m *Main) {
			err := m.editView.SetCommand(panel.EditCommandMode, item.Cmd)
			if err != nil {
				m.logger.Error("error setting edit view content", slog.Any("error", err))
			}
		})
	case key.Matches(msg, m.keys.Delete):
		item, ok := m.commandsView.SelectedItem()
		if !ok {
			break
		}

		if err := m.removeCommand(*item.Cmd); err != nil {
			m.logger.Error("error removing command", slog.Any("command", *item.Cmd), slog.Any("error", err))
		}
	default:
		m.commandsView, cmd = m.commandsView.Update(msg)
		item, ok := m.commandsView.SelectedItem()
		if !ok {
			break
		}

		if !item.Loaded {
			c, err := m.fechFullCommand(item.Cmd.ID.String())
			if err != nil {
				m.logger.Error("error fetching command details", slog.Any("error", err))
				break
			}

			item.Cmd.Params = c.Params
			item.Loaded = true

			m.logger.Debug("command details fetched successfully", slog.Any("command", c))
		}
		if err := m.detailView.SetCommand(*item.Cmd); err != nil {
			m.logger.Error("error setting detail view content", slog.Any("error", err))
		}
	}
	return cmd
}

func (m *Main) handleExecuteInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Back):
		return ChangeFocus(NavigationFocus, func(m *Main) {
			item, ok := m.commandsView.SelectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.Cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.executeView, cmd = m.executeView.Update(msg)
		return cmd
	}
}

func (m *Main) handleEditInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Back):
		return ChangeFocus(NavigationFocus, func(m *Main) {
			item, ok := m.commandsView.SelectedItem()
			if !ok {
				return
			}
			if err := m.detailView.SetCommand(*item.Cmd); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.editView, cmd = m.editView.Update(msg)
		return cmd
	}
}

func (m *Main) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	getAll := func() {
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands", slog.Any("error", err))
			return
		}

		m.setContent(cmds)
	}

	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.Back):
		return ChangeFocus(NavigationFocus, nil)
	case key.Matches(msg, m.keys.DiscardSearch):
		m.searchView.Reset()
		getAll()
		return ChangeFocus(NavigationFocus, nil)
	case key.Matches(msg, m.keys.Enter):
		m.searchView.Unfocus()
		return ChangeFocus(NavigationFocus, nil)
	default:
		m.searchView, cmd = m.searchView.Update(msg)

		terms := m.searchView.Content()
		if len(terms) >= minCharCount {
			m.searching = true
			cmds, err := m.searchCommands(terms)
			if err != nil {
				m.logger.Error("error searching for commands",
					slog.String("terms", terms),
					slog.Any("error", err),
				)
				break
			}
			m.setContent(cmds)
		} else if m.searching && len(terms) == 0 {
			m.searching = false
			getAll()
		}
	}
	return cmd
}
