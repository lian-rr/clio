package view

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/command/manager"
	"github.com/lian-rr/clio/tui/components/dialog"
	"github.com/lian-rr/clio/tui/view/msgs"
	"github.com/lian-rr/clio/tui/view/panel"
)

const minCharCount = 3

func (m *Main) handleInput(msg tea.Msg) tea.Cmd {
	// TODO: this is getting anoying, review this later, consider approach where the handlers are registered and then with a map[focus]handler is chosen.
	handler := func(msg tea.Msg) tea.Cmd {
		switch m.focus {
		case searchFocus:
			return m.handleSearchInput(msg)
		case executeFocus:
			return m.handleExecuteInput(msg)
		case editFocus:
			return m.handleEditInput(msg)
		case explainFocus:
			return m.handleExplainInput(msg)
		case historyFocus:
			return m.handleHistoryInput(msg)
		default:
			return m.handleNavigationInput(msg)
		}
	}
	switch msg := msg.(type) {
	case dialog.InitMsg:
		m.confirmation = true
		m.logger.Debug("dialog open", slog.Bool("confirmation", m.confirmation))
		return handler(msg)
	case dialog.AcceptMsg, dialog.DiscardMsg:
		m.confirmation = false
		m.logger.Debug("dialog closed", slog.Bool("confirmation", m.confirmation))
		return handler(msg)
	default:
		return handler(msg)
	}
}

func (m *Main) handleNavigationInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmation {
			m.detailPanel, cmd = m.detailPanel.Update(msg)
			break
		}

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
		case key.Matches(msg, m.keys.Compose):
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
			m.editPanel.Reset()

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
			if m.professor == nil {
				m.logger.Warn("professor not available")
				break
			}

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
		case key.Matches(msg, m.keys.History):
			item, ok := m.explorerPanel.SelectedCommand()
			if !ok {
				break
			}

			msgs.PublishAsyncMsg(m.activityChan, msgs.HandleRequestHistoryMsg(item.Command.ID))

			return changeFocus(historyFocus, func(m *Main) {
				err := m.historyPanel.SetCommand(*item.Command)
				if err != nil {
					m.logger.Error("error setting history view content", slog.Any("error", err))
				}
			})
		case key.Matches(msg, m.keys.Delete):
			return m.detailPanel.ToggleConfirmation()
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
	// delete command confirmed
	case dialog.AcceptMsg:
		_ = m.detailPanel.ToggleConfirmation()

		item, ok := m.explorerPanel.SelectedCommand()
		if !ok {
			break
		}

		if err := m.removeCommand(*item.Command); err != nil {
			m.logger.Error("error removing command", slog.Any("command", *item.Command), slog.Any("error", err))
		}
	case dialog.DiscardMsg:
		_ = m.detailPanel.ToggleConfirmation()
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
			if m.confirmation {
				break
			}

			return changeFocus(navigationFocus, func(m *Main) {
				item, ok := m.explorerPanel.SelectedCommand()
				if !ok {
					return
				}
				if err := m.detailPanel.SetCommand(*item.Command); err != nil {
					m.logger.Error("error setting detail view content", slog.Any("error", err))
				}
			})
		}
	}

	// pass control for any other event
	m.editPanel, cmd = m.editPanel.Update(msg)
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

	// TODO: change this to follow the aproach of other panels
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
		case key.Matches(msg, m.keys.Go):
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

func (m *Main) handleHistoryInput(msg tea.Msg) tea.Cmd {
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
			m.historyPanel, cmd = m.historyPanel.Update(msg)
		}
	default:
		// pass control for any other event
		m.historyPanel, cmd = m.historyPanel.Update(msg)
	}
	return cmd
}

func (m *Main) handleAsyncActivities(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.RequestExplanationMsg:
		go m.fetchExplanation(msg.Command)
	case msgs.SetExplanationMsg:
		if msg.Cache {
			go m.cacheExplanation(msg.CommandID, msg.Explanation)
		}
		m.explainPanel.SetExplanation(msg.Explanation)
	case msgs.CacheExplanationMsg:
		go m.cacheExplanation(msg.CommandID, msg.Explanation)
	case msgs.EvictCachedExplanationMsg:
		go m.deleteExplanation(msg.CommandID)
	case msgs.RequestHistoryMsg:
		go m.fetchHistory(msg.CommandID)
	case msgs.SetHistoryMsg:
		m.historyPanel.SetHistory(msg.History)
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
	ctx, cancel := context.WithTimeout(m.ctx, time.Second*60)
	defer cancel()

	var cache bool
	explanation, err := m.commandController.ReadExplanation(ctx, cmd.ID)
	if err != nil {
		if !errors.Is(err, manager.ErrElementNotFound) {
			m.logger.Error("error getting command explanation from cache",
				slog.Any("command", "cmd"),
				slog.Any("error", err),
			)
		}

		m.logger.Debug("getting explanation from professor")
		explanation, err = m.professor.Explain(ctx, cmd)
		if err != nil {
			m.logger.Error("error getting command explanation from professor",
				slog.Any("command", "cmd"),
				slog.Any("error", err),
			)
			return
		}

		cache = true
	}

	msgs.PublishAsyncMsg(
		m.activityChan,
		msgs.HandleSetExplanationMsg(cmd.ID, explanation, cache),
	)
}

func (m *Main) cacheExplanation(commandID uuid.UUID, explanation string) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*400)
	defer cancel()

	m.logger.Debug("attempting to cache explanation")
	err := m.commandController.WriteExplanation(ctx, commandID, explanation)
	if err != nil {
		m.logger.Error("error writing explanation in cache", slog.Any("error", err))
	}
}

func (m *Main) deleteExplanation(commandID uuid.UUID) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*400)
	defer cancel()

	m.logger.Debug("attempting to delete explanation cache")
	err := m.commandController.DeleteExplanation(ctx, commandID)
	if err != nil {
		m.logger.Error("error deleting explanation from cache", slog.Any("error", err))
	}
}

func (m *Main) fetchHistory(commandID uuid.UUID) {
	// ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*400)
	// defer cancel()

	history := command.History{
		Usages: []command.Usage{
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
			{
				Command:   "test 1",
				Timestamp: time.Now(),
			},
		},
	}

	msgs.PublishAsyncMsg(
		m.activityChan,
		msgs.HandleSetHistoryMsg(history),
	)
}
