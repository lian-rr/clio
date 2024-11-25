package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

// App holds the application configuration
type App struct {
	BasePath  string
	Professor ProfessorConfig
}

// ErrInvalidDir indicates that the root directory for the store is not valid.
var ErrInvalidDir = errors.New("invalid base directory")

// New returns a new app.
func New(ctx context.Context, path string, logger *slog.Logger) (App, error) {
	if path == "" {
		def, err := os.UserHomeDir()
		if err != nil {
			return App{}, err
		}
		path = def
	}

	path = fmt.Sprintf("%s/.clio", path)
	err := os.Mkdir(path, 0o740)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return App{}, fmt.Errorf("%w: %w", ErrInvalidDir, err)
	}

	app := App{
		BasePath: path,
	}

	if err := app.loadProfessor(); err != nil {
		logger.Error("error loading professor config", slog.Any("error", err))
	}

	logger.Info("application setup completed", slog.String("path", path))
	return app, nil
}
