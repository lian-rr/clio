package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

// ErrNoConfigFound config file not found
var ErrNoConfigFound = errors.New("config file or not found")

// App holds the application configuration
type App struct {
	PathOverride string `toml:"pathOverride"`
	Debug        bool   `toml:"debug"`
	Professor    ProfessorConfig
}

// New returns a new app's config.
func New(configPath string) (App, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return App{}, err
		}
		configPath = fmt.Sprintf("%s/.config/clio", home)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return App{}, err
	}

	if err := cfg.prepateDataDir(); err != nil {
		return App{}, err
	}

	if err := cfg.validate(); err != nil {
		return App{}, err
	}

	return cfg, nil
}

// NewDefault returns a default App.
func NewDefault() (App, error) {
	var cfg App
	err := cfg.prepateDataDir()
	if err != nil {
		return App{}, err
	}

	return cfg, nil
}

func loadConfig(path string) (App, error) {
	var cfg App
	_, err := toml.DecodeFile(fmt.Sprintf("%s/clio.toml", path), &cfg)
	if err != nil {
		return App{}, fmt.Errorf("%w path=%q: %v", ErrNoConfigFound, path, err)
	}
	return cfg, nil
}

func (a App) prepateDataDir() error {
	path := a.GetPath()
	err := os.Mkdir(path, 0o740)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("error preparing the data dir (path=%s): %v", path, err)
	}

	return nil
}

func (a App) validate() error {
	var errs error

	if err := a.Professor.validate(); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func (a App) GetPath() string {
	path := a.PathOverride
	if path == "" {
		path, _ = os.UserHomeDir()
	}

	return fmt.Sprintf("%s/.clio", path)
}
