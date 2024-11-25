package teacher

import (
	"context"
	"errors"
	"log/slog"

	"github.com/lian-rr/clio/command"
)

var ErrSourceNotSet error = errors.New("source not set")

type Source interface {
	Prompt(context.Context, string) (string, error)
}

type OptFunc func(teacher *Teacher)

type Teacher struct {
	source Source
	logger *slog.Logger
}

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
