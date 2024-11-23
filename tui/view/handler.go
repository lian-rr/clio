package view

import (
	"log/slog"
	"reflect"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/panel"
)

const minCharCount = 3

func (m *Main) handleInput(msg tea.Msg) tea.Cmd {
	switch m.focus {
	case searchFocus:
		return m.handleSearchInput(msg)
	case executeFocus:
		return m.handleExecuteInput(msg)
	case editFocus:
		return m.handleEditInput(msg)
	case explainFocus:
		return m.handleExplainInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}
}

func (m *Main) handleNavigationInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return tea.Quit
		case key.Matches(msg, m.keys.Search):
			return changeFocus(searchFocus, func(m *Main) {
				m.searchPanel.Focus()
			})
		case key.Matches(msg, m.keys.DiscardSearch):
			m.searchPanel.Reset()
			cmds, err := m.fechCommands()
			if err != nil {
				m.logger.Error("error getting all commands", slog.Any("error", err))
				break
			}

			m.setContent(cmds)
		case key.Matches(msg, m.keys.Enter):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			return changeFocus(executeFocus, func(m *Main) {
				err := m.executePanel.SetCommand(*item.Command)
				if err != nil {
					m.logger.Error("error setting execute view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.New):
			return changeFocus(editFocus, func(m *Main) {
				err := m.editPanel.SetCommand(panel.NewCommandMode, nil)
				if err != nil {
					m.logger.Error("error setting edit view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.Edit):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			return changeFocus(editFocus, func(m *Main) {
				err := m.editPanel.SetCommand(panel.EditCommandMode, item.Command)
				if err != nil {
					m.logger.Error("error setting edit view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.Copy):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			return changeFocus(editFocus, func(m *Main) {
				err := m.editPanel.SetCommand(panel.NewCommandMode, item.Command)
				if err != nil {
					m.logger.Error("error setting edit view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.Explain):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			msgs.PublishAsyncMsg(m.activityChan, msgs.HandleRequestExplanationMsg(*item.Command))

			return changeFocus(explainFocus, func(m *Main) {
				err := m.explainPanel.SetCommand(*item.Command)
				if err != nil {
					m.logger.Error("error setting explain view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.Delete):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			if err := m.removeCommand(*item.Command); err != nil {
				m.logger.Error("error removing command", slog.Any("command", *item.Command), slog.Any("error", err))
			}
		default:
			m.explorerPanel, cmd = m.explorerPanel.Update(msg)
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			if !item.Loaded {
				c, err := m.fechFullCommand(item.Command.ID.String())
				if err != nil {
					m.logger.Error("error fetching command details", slog.Any("error", err))
					break
				}

				item.Command.Params = c.Params
				item.Loaded = true

				m.logger.Debug("command details fetched successfully", slog.Any("command", c))
			}
			if err := m.detailPanel.SetCommand(*item.Command); err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		}
	}

	return cmd
}

func (m *Main) handleExecuteInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return changeFocus(navigationFocus, func(m *Main) {
				item, ok := m.explorerPanel.SelectedCommand()
				if !ok {
					return
				}
				if err := m.detailPanel.SetCommand(*item.Command); err != nil {
					m.logger.Error("error setting detail view content", slog.Any("error", err))
				}
			})
		default:
			m.executePanel, cmd = m.executePanel.Update(msg)
		}
	default:
		// pass control for any other event
		m.executePanel, cmd = m.executePanel.Update(msg)
	}
	return cmd
}

func (m *Main) handleEditInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return changeFocus(navigationFocus, func(m *Main) {
				item, ok := m.explorerPanel.SelectedCommand()
				if !ok {
					return
				}
				if err := m.detailPanel.SetCommand(*item.Command); err != nil {
					m.logger.Error("error setting detail view content", slog.Any("error", err))
				}
			})
		default:
			m.editPanel, cmd = m.editPanel.Update(msg)
		}
	default:
		// pass control for any other event
		m.editPanel, cmd = m.editPanel.Update(msg)
	}
	return cmd
}

func (m *Main) handleSearchInput(msg tea.Msg) tea.Cmd {
	getAll := func() {
		cmds, err := m.fechCommands()
		if err != nil {
			m.logger.Error("error getting all commands", slog.Any("error", err))
			return
		}

		m.setContent(cmds)
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.searchPanel.Unfocus()
			return changeFocus(navigationFocus, nil)
		case key.Matches(msg, m.keys.DiscardSearch):
			m.searchPanel.Reset()
			getAll()
			return changeFocus(navigationFocus, nil)
		case key.Matches(msg, m.keys.Enter):
			m.searchPanel.Unfocus()
			return changeFocus(navigationFocus, nil)
		default:
			m.searchPanel, cmd = m.searchPanel.Update(msg)

			terms := m.searchPanel.Content()
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
				if len(cmds) > 0 {
					m.explorerPanel.Select(0)
				}
			} else if m.searching && len(terms) == 0 {
				m.searching = false
				getAll()
			}
		}
	default:
		// pass control for any other event
		m.searchPanel, cmd = m.searchPanel.Update(msg)
	}
	return cmd
}

func (m *Main) handleExplainInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return changeFocus(navigationFocus, func(m *Main) {
				item, ok := m.explorerPanel.SelectedCommand()
				if !ok {
					return
				}
				if err := m.detailPanel.SetCommand(*item.Command); err != nil {
					m.logger.Error("error setting detail view content", slog.Any("error", err))
				}
			})
		default:
			m.explainPanel, cmd = m.explainPanel.Update(msg)
		}
	default:
		// pass control for any other event
		m.explainPanel, cmd = m.explainPanel.Update(msg)
	}
	return cmd
}

func (m *Main) handleAsyncActivities(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.RequestExplanationMsg:
		go m.fetchExplanation(msg.Command)
	case msgs.SetExplanationMsg:
		m.explainPanel.SetExplanation(msg.Explanation)
	default:
		m.logger.Warn("unknown async msg captured",
			slog.Any("msg", msg),
			slog.Any("type", reflect.TypeOf(msg)),
		)
	}

	// restart event loop
	return msgs.AsyncHandler(m.activityChan)
}

func (m *Main) fetchExplanation(cmd command.Command) {
	time.Sleep(time.Second * 2)
	msgs.PublishAsyncMsg(
		m.activityChan,
		msgs.HandleSetExplanationMsg(content),
	)
}
