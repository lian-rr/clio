package manager

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/command/sql"
)

var (
	// ErrNotebookNotEnabled thrown when the notebook wasn't set.
	ErrNotebookNotEnabled error = errors.New("notebook not enabled")
	// ErrElementNotFound thrown when the element was not found in the store.
	ErrElementNotFound error = errors.New("element not found")
)

type store interface {
	Save(context.Context, command.Command) error
	GetCommandByID(context.Context, uuid.UUID) (command.Command, error)
	SearchCommand(context.Context, string) ([]command.Command, error)
	ListCommands(context.Context) ([]command.Command, error)
	DeleteCommand(context.Context, uuid.UUID) error
	DeleteParameters(context.Context, []uuid.UUID) error
}

type notebook interface {
	WriteExplanation(context.Context, uuid.UUID, string) error
	ReadExplanation(context.Context, uuid.UUID) (string, error)
	DeleteExplanation(context.Context, uuid.UUID) error
}

// Manager handles the command admin operations.
type Manager struct {
	store    store
	notebook notebook
}

// NewManager returns a new Manager.
func NewManager(store store, notebook notebook) (Manager, error) {
	if store == nil {
		return Manager{}, errors.New("nil store")
	}
	return Manager{
		store:    store,
		notebook: notebook,
	}, nil
}

// Add creates, saves and returns a new command validated.
func (m *Manager) Add(ctx context.Context, cmd command.Command) (command.Command, error) {
	var err error
	cmd.ID, err = uuid.NewV7()
	if err != nil {
		return command.Command{}, nil
	}

	if err := m.store.Save(ctx, cmd); err != nil {
		return command.Command{}, err
	}

	return cmd, nil
}

// GetCommand returns a command by ID.
func (m *Manager) GetOne(ctx context.Context, rawID string) (command.Command, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return command.Command{}, err
	}

	cmd, err := m.store.GetCommandByID(ctx, id)
	if err != nil {
		return command.Command{}, err
	}

	return cmd, nil
}

// SearchCommand returns a list of commands with a matching term.
func (m *Manager) Search(ctx context.Context, term string) ([]command.Command, error) {
	commands, err := m.store.SearchCommand(ctx, term)
	if err != nil {
		return nil, err
	}

	return commands, nil
}

// GetAll returns a list with all the commands.
func (m *Manager) GetAll(ctx context.Context) ([]command.Command, error) {
	commands, err := m.store.ListCommands(ctx)
	if err != nil {
		return nil, err
	}

	return commands, nil
}

// GetCommand returns a command by ID.
func (m *Manager) DeleteCommand(ctx context.Context, rawID string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return err
	}

	return m.store.DeleteCommand(ctx, id)
}

// UpdateCommand updates the command on the store.
func (m *Manager) UpdateCommand(ctx context.Context, cmd command.Command) (command.Command, error) {
	curr, err := m.store.GetCommandByID(ctx, cmd.ID)
	if err != nil {
		return command.Command{}, fmt.Errorf("error getting current command: %v", err)
	}

	toDeleteParams := make([]uuid.UUID, 0)
	for _, param := range curr.Params {
		var found bool
		for _, newp := range cmd.Params {
			if param.ID == newp.ID {
				found = true
				break
			}
		}

		if !found {
			toDeleteParams = append(toDeleteParams, param.ID)
		}
	}

	if err := m.store.Save(ctx, cmd); err != nil {
		return command.Command{}, fmt.Errorf("error updating command: %v", err)
	}

	if err := m.store.DeleteParameters(ctx, toDeleteParams); err != nil {
		return command.Command{}, err
	}
	return cmd, nil
}

// WriteExplanation writes the explanation in the notebook.
func (m *Manager) WriteExplanation(ctx context.Context, commandID uuid.UUID, explanation string) error {
	if m.notebook == nil {
		return ErrNotebookNotEnabled
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("error creating writer: %v", err)
	}

	_, err = writer.Write([]byte(explanation))
	if err != nil {
		return fmt.Errorf("error compressing explanation: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error clossing writer: %v", err)
	}

	text := base64.StdEncoding.EncodeToString(buf.Bytes())
	return m.notebook.WriteExplanation(ctx, commandID, text)
}

// ReadExplanation reads the explanation from the notebook.
func (m *Manager) ReadExplanation(ctx context.Context, commandID uuid.UUID) (string, error) {
	if m.notebook == nil {
		return "", ErrNotebookNotEnabled
	}

	encoded, err := m.notebook.ReadExplanation(ctx, commandID)
	if err != nil {
		if errors.Is(err, sql.ErrNotFound) {
			return "", ErrElementNotFound
		}
		return "", err
	}

	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("error base64 decoding explanation: %v", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader([]byte(compressed)))
	if err != nil {
		return "", fmt.Errorf("error creating reader: %v", err)
	}

	defer func() {
		_ = reader.Close()
	}()

	explanation, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(explanation), nil
}

// DeleteExplanation deletes the explanation from the notebook.
func (m *Manager) DeleteExplanation(ctx context.Context, commandID uuid.UUID) error {
	if m.notebook == nil {
		return ErrNotebookNotEnabled
	}

	if err := m.notebook.DeleteExplanation(ctx, commandID); err != nil {
		if !errors.Is(err, sql.ErrNotFound) {
			return err
		}
	}

	return nil
}
