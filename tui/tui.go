package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian-rr/clio/command/manager"
	"github.com/lian-rr/clio/command/professor"
	"github.com/lian-rr/clio/out"
	"github.com/lian-rr/clio/tui/view"
)

// Tui contains the TUI logic.
type Tui struct {
	program *tea.Program
	logger  *slog.Logger
}

// New returns a new TUI container.
func New(ctx context.Context, manager *manager.Manager, logger *slog.Logger, professor *professor.Professor) (Tui, error) {
	opts := make([]view.OptFunc, 0)
	if professor != nil {
		opts = append(opts, view.WithProfessor(professor))
	}

	model, err := view.New(ctx, manager, logger, opts...)
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

	mm, ok := m.(*view.Main)
	if !ok {
		return errors.New("error getting last model")
	}
	if mm.Output != "" {
		t.logger.Debug("program output", slog.String("command", mm.Output))
		out.Produce(mm.Output)
		out.Clear()
	}

	return nil
}
