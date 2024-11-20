package command

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

var regex = regexp.MustCompile(`{{\s?\.\w+\s?}}`)

// ErrInvalidNumOfParams returned when the number of params provided doesn't match the command
var ErrInvalidNumOfParams = errors.New("invalid number of params provided")

type (
	// Command represents a shell.
	Command struct {
		ID          uuid.UUID
		Name        string
		Description string
		Command     string
		Params      []Parameter
	}

	// Parameter represents the Command Parameter
	Parameter struct {
		ID           uuid.UUID
		Name         string
		Description  string
		DefaultValue string
	}
	// Argument represents the command arguments to place in the params
	Argument struct {
		Name  string
		Value string
	}
)

type cmdOpt func(*Command) error

// New returns a new Command.
func New(name string, desc string, rawCmd string, opts ...cmdOpt) (Command, error) {
	cmd := Command{
		Name:        name,
		Description: desc,
		Command:     rawCmd,
	}

	for _, opt := range opts {
		if err := opt(&cmd); err != nil {
			return Command{}, err
		}
	}

	if err := cmd.Build(); err != nil {
		return Command{}, err
	}

	return cmd, nil
}

// Build builds the internal attributes (template and params).
func (c *Command) Build() error {
	if c.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		c.ID = id
	}

	news := parseParams(c.Command)
	params := make([]Parameter, 0, len(news))
	for _, param := range news {
		for j := 0; j < len(c.Params); j++ {
			old := c.Params[j]
			if param.Name == old.Name {
				param.ID = old.ID
				param.Description = old.Description
				param.DefaultValue = old.DefaultValue
				break
			}
		}
		params = append(params, param)
	}

	c.Params = params
	return nil
}

// Compile returns the command with the arguments applied.
func (c *Command) Compile(args []Argument) (string, error) {
	tmpl, err := template.New(c.Name).Parse(c.Command)
	if err != nil {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	if len(args) != len(c.Params) {
		return "", ErrInvalidNumOfParams
	}

	arguments := make(map[string]string, len(args))
	for _, arg := range args {
		arguments[arg.Name] = arg.Value
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, arguments); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func parseParams(raw string) []Parameter {
	rawParams := regex.FindAllString(raw, -1)

	params := make([]Parameter, 0, len(rawParams))
	for _, rp := range rawParams {
		id, _ := uuid.NewV6()
		param := Parameter{
			ID:   id,
			Name: rp[3 : len(rp)-2],
		}

		params = append(params, param)
	}

	return params
}

// WithParams used to pass the params to the Command.
// Returns an error if the param is not found.
func WithParams(params []Parameter) cmdOpt {
	return func(c *Command) error {
		for _, param := range params {
			if !strings.Contains(c.Command, fmt.Sprintf("{{.%s}}", param.Name)) {
				return fmt.Errorf("param '%s' not found in the command", param.Name)
			}

			c.Params = append(c.Params, param)
		}

		return nil
	}
}
