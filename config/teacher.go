package config

import (
	"errors"
	"os"
	"strconv"
)

type TeacherType string

const (
	OpenAITeacher TeacherType = "openai"
)

type TeacherConfig struct {
	Enabled bool
	Type    TeacherType
	OpenAI  OpenAIConfig
}

type OpenAIConfig struct {
	ApiKey       string
	CustomPrompt string
	Url          string
	Model        string
}

func (a *App) loadTeacher() error {
	cfg := TeacherConfig{}
	if str, ok := os.LookupEnv("TEACHER_ENABLED"); ok {
		val, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		cfg.Enabled = val
	}

	if str, ok := os.LookupEnv("TEACHER_TYPE"); ok {
		val, err := toTeacherType(str)
		if err != nil {
			return err
		}
		cfg.Type = val
	} else {
		return errors.New("teacher type not set")
	}

	if cfg.Type == OpenAITeacher {
		err := cfg.loadOpenAIConfig()
		if err != nil {
			return err
		}
	}

	a.Teacher = cfg
	return nil
}

func toTeacherType(str string) (TeacherType, error) {
	switch str {
	case "openai":
		return OpenAITeacher, nil
	default:
		return "", errors.New("invalid teacher type")
	}
}

func (tcfg *TeacherConfig) loadOpenAIConfig() error {
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
