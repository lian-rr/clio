package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/keep/command"
	"github.com/lian-rr/keep/out"
)

type manager interface {
	GetAll(context.Context) ([]command.Command, error)
	GetOne(context.Context, string) (command.Command, error)
	Search(context.Context, string) ([]command.Command, error)
	Add(context.Context, command.Command) (command.Command, error)
	DeleteCommand(context.Context, string) error
	UpdateCommand(context.Context, command.Command) (command.Command, error)
}

// Tui contains the TUI logic.
type Tui struct {
	program *tea.Program
	logger  *slog.Logger
}

// New returns a new TUI container.
func New(ctx context.Context, manager *command.Manager, logger *slog.Logger) (Tui, error) {
	model, err := newMain(ctx, manager, logger)
	if err != nil {
		return Tui{}, fmt.Errorf("error starting the main model: %w", err)
	}

	return Tui{
		program: tea.NewProgram(
			model,
			tea.WithContext(ctx),
			tea.WithAltScreen(),
		),
		logger: logger,
	}, nil
}

// Start start the TUI app.
func (t *Tui) Start() error {
	m, err := t.program.Run()
	if err != nil {
		return fmt.Errorf("error starting the TUI program: %w", err)
	}

	mm, ok := m.(*main)
	if !ok {
		return errors.New("error getting last model")
	}
	if mm.output != "" {
		t.logger.Debug("program output", slog.String("command", mm.output))
		out.Produce(mm.output)
		out.Clear()
	}

	return nil
}
