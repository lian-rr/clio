package config

import (
	"errors"
)

// ProfessorSourceType type for the professor source implementation.
type ProfessorSourceType string

const (
	// OpenAISourceType used for setting OpenAI as the professor source.
	OpenAISourceType ProfessorSourceType = "openai"
)

// ProfessorConfig is the config for the Professor feature.
type ProfessorConfig struct {
	Enabled bool                `toml:"enabled"`
	Type    ProfessorSourceType `toml:"type"`
	OpenAI  OpenAISourceConfig  `toml:"openai"`
}

// OpenAISourceConfig holds the configuration for setting the OpenAI professor source.
type OpenAISourceConfig struct {
	ApiKey       string `toml:"key"`
	CustomPrompt string `toml:"customPrompt"`
	Url          string `toml:"url"`
	Model        string `toml:"model"`
}

func (p ProfessorConfig) validate() error {
	if !p.Enabled {
		return nil
	}

	switch p.Type {
	case OpenAISourceType:
		return p.OpenAI.validate()
	default:
		return errors.New("invalid professor type")
	}
}

func (c OpenAISourceConfig) validate() error {
	if c.ApiKey == "" {
		return errors.New("missing openai api key")
	}

	return nil
}
