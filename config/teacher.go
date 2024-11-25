package config

import (
	"errors"
	"os"
	"strconv"
)

// ProfessorType type for the professor implementation.
type ProfessorType string

const (
	// OpenAIProfessor used for setting OpenAI as the professor source.
	OpenAIProfessor ProfessorType = "openai"
)

// ProfessorConfig is the config for the Professor feature.
type ProfessorConfig struct {
	Enabled bool
	Type    ProfessorType
	OpenAI  OpenAIConfig
}

// OpenAIConfig holds the configuration for setting the OpenAI professor source.
type OpenAIConfig struct {
	ApiKey       string
	CustomPrompt string
	Url          string
	Model        string
}

func (a *App) loadProfessor() error {
	cfg := ProfessorConfig{}
	if str, ok := os.LookupEnv("PROFE_ENABLED"); ok {
		val, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		cfg.Enabled = val
	}

	if str, ok := os.LookupEnv("PROFE_TYPE"); ok {
		val, err := toProfessorType(str)
		if err != nil {
			return err
		}
		cfg.Type = val
	} else {
		return errors.New("professor type not set")
	}

	if cfg.Type == OpenAIProfessor {
		err := cfg.loadOpenAIConfig()
		if err != nil {
			return err
		}
	}

	a.Professor = cfg
	return nil
}

func toProfessorType(str string) (ProfessorType, error) {
	switch str {
	case "openai":
		return OpenAIProfessor, nil
	default:
		return "", errors.New("invalid professor type")
	}
}

func (tcfg *ProfessorConfig) loadOpenAIConfig() error {
	cfg := OpenAIConfig{}

	if str, ok := os.LookupEnv("OPENAI_KEY"); ok {
		cfg.ApiKey = str
	}

	if cfg.ApiKey == "" {
		return errors.New("missing openAI api key")
	}

	if str, ok := os.LookupEnv("OPENAI_URL"); ok {
		cfg.Url = str
	}

	if str, ok := os.LookupEnv("OPENAI_MODEL"); ok {
		cfg.Model = str
	}

	if str, ok := os.LookupEnv("OPENAI_PROMPT"); ok {
		cfg.CustomPrompt = str
	}

	tcfg.OpenAI = cfg
	return nil
}
