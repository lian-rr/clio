package teacher

import (
	"context"
	"errors"
	"log/slog"

	"github.com/lian-rr/clio/command"
)

// ErrSourceNotSet thrown when the source is not set.
var ErrSourceNotSet error = errors.New("source not set")

// Source is the source of information for the teacher.
type Source interface {
	Prompt(context.Context, string) (string, error)
}

// OptFunc used for setting optional configs.
type OptFunc func(teacher *Teacher)

// Teacher handles the explanation of commands.
type Teacher struct {
	source Source
	logger *slog.Logger
}

// New returns a new teacher.
func New(source Source, logger *slog.Logger, opts ...OptFunc) Teacher {
	teach := Teacher{
		source: source,
		logger: logger,
	}

	for _, opt := range opts {
		opt(&teach)
	}

	return teach
}

// Explain the passed command. If the source is not set, then it will return ErrSourceNotSet.
func (t Teacher) Explain(ctx context.Context, cmd command.Command) (string, error) {
	if t.source == nil {
		return "", ErrSourceNotSet
	}

	resp, err := t.source.Prompt(ctx, cmd.Command)
	if err != nil {
		return "", err
	}

	return resp, nil
}
