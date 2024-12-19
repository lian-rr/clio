package view

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
)

type controller interface {
	GetAll(context.Context) ([]command.Command, error)
	GetOne(context.Context, string) (command.Command, error)
	Search(context.Context, string) ([]command.Command, error)
	Add(context.Context, command.Command) (command.Command, error)
	DeleteCommand(context.Context, string) error
	UpdateCommand(context.Context, command.Command) (command.Command, error)
	WriteExplanation(context.Context, uuid.UUID, string) error
	ReadExplanation(context.Context, uuid.UUID) (string, error)
	DeleteExplanation(context.Context, uuid.UUID) error
}

func (m *Main) fechCommands() ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandController.GetAll(ctx)
}

func (m *Main) searchCommands(term string) ([]command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*300)
	defer cancel()

	return m.commandController.Search(ctx, term)
}

func (m *Main) fechFullCommand(id string) (command.Command, error) {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	return m.commandController.GetOne(ctx, id)
}

func (m *Main) saveCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	cmd, err := m.commandController.Add(ctx, cmd)
	if err != nil {
		return err
	}

	idx := m.explorerPanel.AddCommand(cmd)
	m.explorerPanel.Select(idx)
	m.detailPanel.SetCommand(cmd)

	return nil
}

func (m *Main) editCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	newCmd, err := m.commandController.UpdateCommand(ctx, cmd)
	if err != nil {
		return err
	}

	m.explorerPanel.RefreshCommand(newCmd)
	m.detailPanel.SetCommand(newCmd)
	return nil
}

func (m *Main) removeCommand(cmd command.Command) error {
	ctx, cancel := context.WithTimeout(m.ctx, time.Millisecond*200)
	defer cancel()

	err := m.commandController.DeleteCommand(ctx, cmd.ID.String())
	if err != nil {
		return err
	}

	toSelectPos := m.explorerPanel.RemoveSelectedCommand()
	if toSelectPos >= 0 {
		m.explorerPanel.Select(toSelectPos)
		if item, ok := m.explorerPanel.SelectedCommand(); ok {
			m.detailPanel.SetCommand(*item.Command)
		}
	} else {
		m.detailPanel.SetCommand(command.Command{})
	}

	return nil
}
