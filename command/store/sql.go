package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/lian_rr/keep/command"
	"github.com/lian_rr/keep/command/store/sqlite"
)

// ErrNotFound used when the searched element wasn't found.
var ErrNotFound = errors.New("not found")

// SqlOptFunc optional functions for Sql store.
type SqlOptFunc func(store *Sql) error

// Sql store
type Sql struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSql returns a new SQL store.
func NewSql(logger *slog.Logger, opts ...SqlOptFunc) (*Sql, error) {
	store := &Sql{
		logger: logger,
	}

	for _, opt := range opts {
		if err := opt(store); err != nil {
			return nil, err
		}
	}

	if store.db == nil {
		return nil, errors.New("missing db connection")
	}

	return store, nil
}

// Save stores a command on the sql store.
func (s *Sql) Save(ctx context.Context, cmd command.Command) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Warn("error rolling back store command", slog.Any("error", err))
			return
		}
	}()

	_, err = tx.ExecContext(ctx, sqlite.InsertCommandQuery, cmd.ID.String(), cmd.Name, cmd.Description, cmd.Command)
	if err != nil {
		return fmt.Errorf("error storing command: %w", err)
	}

	if len(cmd.Params) > 0 {
		placeholders := make([]string, 0, len(cmd.Params))
		args := make([]any, 0, len(cmd.Params)*5) // cap: number of params * attrs to store

		for _, param := range cmd.Params {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args, param.ID.String(), cmd.ID.String(), param.Name, param.Description, param.DefaultValue)
		}

		paramsQuery := fmt.Sprintf(sqlite.InsertParameterPartialQuery, strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, paramsQuery, args...)
		if err != nil {
			return fmt.Errorf("error storing parameters: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}

	s.logger.Debug("command stored successfully")
	return nil
}

// ListCommands returns a list of all the commands
func (s *Sql) ListCommands(ctx context.Context) ([]command.Command, error) {
	rows, err := s.db.QueryContext(ctx, sqlite.GetAllCommandsQuery)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Warn("error closing rows when listing commands", slog.Any("error", err))
		}
	}()

	cmds := make([]command.Command, 0)
	for rows.Next() {
		var cmd command.Command
		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Description, &cmd.Command); err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// GetCommandByID returns a command. If the command doesn't exists, returns an ErrNotFound error.
func (s *Sql) GetCommandByID(ctx context.Context, id uuid.UUID) (command.Command, error) {
	row := s.db.QueryRowContext(ctx, sqlite.GetCommandbyIDQuery, id.String())
	var cmd command.Command
	if err := row.Scan(&cmd.ID, &cmd.Name, &cmd.Description, &cmd.Command); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return command.Command{}, ErrNotFound
		}
		return command.Command{}, err
	}

	rows, err := s.db.QueryContext(ctx, sqlite.GetParametersByCommandID, id.String())
	if err != nil {
		return command.Command{}, err
	}

	params := make([]command.Parameter, 0)
	for rows.Next() {
		var param command.Parameter
		if err := rows.Scan(&param.ID, &param.Name, &param.Description, &param.DefaultValue); err != nil {
			return command.Command{}, err
		}

		params = append(params, param)
	}

	cmd.Params = params
	return cmd, nil
}

// SearchCommand returns a list of the commands with the matching term.
func (s *Sql) SearchCommand(ctx context.Context, term string) ([]command.Command, error) {
	rows, err := s.db.QueryContext(ctx, sqlite.SearchCommandQuery, fmt.Sprintf("%s*", term))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Warn("error closing rows when listing commands", slog.Any("error", err))
		}
	}()

	cmds := make([]command.Command, 0)
	for rows.Next() {
		var cmd command.Command
		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Description, &cmd.Command); err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// DeleteCommand removes a command and it's params.
func (s *Sql) DeleteCommand(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, sqlite.DeleteCommand, id.String())
	if err != nil {
		return fmt.Errorf("error removing command with ID %q: %w", id, err)
	}
	return nil
}

// Close closes the db driver.
func (s *Sql) Close() error {
	return s.db.Close()
}

// WithSqliteDriver returns a sqlOptFunc that sets the config necessary for a SQLite store.
func WithSqliteDriver(ctx context.Context, path string) SqlOptFunc {
	return func(store *Sql) error {
		if store.db != nil {
			return errors.New("sql store already set")
		}

		db, err := sql.Open("sqlite3", fmt.Sprintf("%s/keep.db", path))
		if err != nil {
			return fmt.Errorf("error opening sqlite db: %v", err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		defer func() {
			err := tx.Rollback()
			if !errors.Is(err, sql.ErrTxDone) {
				store.logger.Warn("error rolling back init sqlite db config", slog.Any("error", err))
			}
		}()

		queries := []string{
			sqlite.CommandTableQuery,
			sqlite.ParametersTableQuery,
			sqlite.SearchTableQuery,
			sqlite.InsertCommandFtsTrigger,
			sqlite.UpdateCommandFtsTrigger,
			sqlite.DeleteCommandFtsTrigger,
			// testCommands,
			// testParams,
		}

		for _, query := range queries {
			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("error executing the query `%s`: %w", query, err)
			}
			store.logger.Debug("query executed successfully", slog.String("query", query))
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error commiting transaction: %w", err)
		}

		store.logger.Debug("sqlite store initiatied successfully")

		store.db = db
		return nil
	}
}

var testCommands = `
	INSERT INTO commands (id, name, description, command) VALUES 
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a5', 'CreateUser', 'Creates a new user in the system', 'useradd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a6', 'DeleteUser', 'Removes a user from the system', 'userdel {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a7', 'UpdateUser', 'Updates user information', 'usermod -c "{{.info}}" {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a8', 'ListUsers', 'Retrieves a list of all users', 'cat /etc/passwd | grep {{.filter}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a9', 'GetUser', 'Fetches details of a specific user', 'getent passwd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537aa', 'ChangePassword', 'Updates the password for a user', 'passwd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ab', 'LockUser', 'Locks a user account to prevent access', 'usermod -L {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ac', 'UnlockUser', 'Unlocks a previously locked user account', 'usermod -U {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ad', 'GetLogs', 'Retrieves system logs for audit purposes', 'journalctl -xe --user {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ae', 'SystemStatus', 'Checks the current status of the system', 'systemctl status {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537af', 'BackupDatabase', 'Backs up the specified database', 'pg_dump {{.database}} -f {{.output}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b0', 'RestoreDatabase', 'Restores the specified database from a backup', 'psql {{.database}} < {{.backup}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b1', 'CheckDiskSpace', 'Checks the disk space usage', 'df -h {{.path}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b2', 'StartService', 'Starts a specified service', 'systemctl start {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b3', 'StopService', 'Stops a specified service', 'systemctl stop {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b4', 'CheckServiceStatus', 'Checks the status of a specified service', 'systemctl status {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b5', 'CreateDirectory', 'Creates a new directory', 'mkdir -p {{.path}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b6', 'DeleteFile', 'Deletes a specified file', 'rm -f {{.filename}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b7', 'MoveFile', 'Moves a file to a new location', 'mv {{.source}} {{.destination}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b8', 'CopyFile', 'Copies a file to a new location', 'cp {{.source}} {{.destination}}')
`

var testParams = `
INSERT INTO parameters (id, command, name, description, value) VALUES
-- Parameters for CreateUser
('c9b073b2-982f-77fa-a052-bbc5cfaf29d1', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a5', 'username', 'The username for the new user', '{{.username}}'),

-- Parameters for DeleteUser
('3ad0860b-8f8b-74d0-9534-e3d2bde019f1', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a6', 'username', 'The username of the user to delete', '{{.username}}'),

-- Parameters for UpdateUser
('9c6b4b2b-9c25-77fa-a399-08d0e4d8e39f', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a7', 'info', 'Additional information for the user', '{{.info}}'),
('cb67f939-913b-7129-a0da-4bb90f689b87', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a7', 'username', 'The username of the user to modify', '{{.username}}'),

-- Parameters for ListUsers
('b18eaf53-b29c-710f-9a8d-b9f66fe8cd97', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a8', 'filter', 'A filter for listing users', '{{.filter}}'),

-- Parameters for GetUser
('4cd6ec2c-a78c-79c0-9f66-1e338506bf8e', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a9', 'username', 'The username of the user to retrieve', '{{.username}}'),

-- Parameters for ChangePassword
('e80f7327-9b01-7b4d-a35b-0bfa460c6d97', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537aa', 'username', 'The username of the user to change the password for', '{{.username}}'),

-- Parameters for LockUser
('3f5cfa56-b4d2-7467-b0b1-8ff28b8b576f', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ab', 'username', 'The username of the user to lock', '{{.username}}'),

-- Parameters for UnlockUser
('d62ad079-d722-7b3a-bc98-2db971a0b0e6', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ac', 'username', 'The username of the user to unlock', '{{.username}}'),

-- Parameters for GetLogs
('5adf91b7-c0e2-7398-bce2-5779d609cae3', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ad', 'username', 'The username whose logs to fetch', '{{.username}}'),

-- Parameters for SystemStatus
('6b27d379-99b1-7058-b57b-5a6d2b849f4f', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ae', 'service', 'The service to check the status of', ''),

-- Parameters for BackupDatabase
('17792b46-bba3-7420-bb35-c6d9b9fa073d', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537af', 'database', 'The name of the database to backup', '{{.database}}'),
('438b49d9-8e61-742d-b319-451da96e1c59', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537af', 'output', 'The output file for the database backup', '{{.output}}'),

-- Parameters for RestoreDatabase
('bc56918d-9b51-72ad-90e9-14c104d3d0cf', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b0', 'database', 'The name of the database to restore', '{{.database}}'),
('4b3f06c0-d24c-7131-9119-f173c5a0b9ba', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b0', 'backup', 'The backup file to restore from', '{{.backup}}'),

-- Parameters for CheckDiskSpace
('789d4e7b-b358-74f1-b02e-bb40b5c1b32c', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b1', 'path', 'The directory or mount point to check disk space for', '{{.path}}'),

-- Parameters for StartService
('c498a16d-44e3-70ca-bb83-94fbb8e6bcdd', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b2', 'service', 'The service to start', ''),

-- Parameters for StopService
('ad10d4b5-84fc-75a2-b672-3e38ef7a1e5f', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b3', 'service', 'The service to stop', ''),

-- Parameters for CheckServiceStatus
('05a6d707-b39e-77c8-a104-f5a230d0a3f3', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b4', 'service', 'The service to check the status of', ''),

-- Parameters for CreateDirectory
('178a56d2-b730-775b-82f4-e7bc430f39b2', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b5', 'path', 'The path for the directory to create', '{{.path}}'),

-- Parameters for DeleteFile
('a0d8d8bc-bda2-76d9-b074-b0a8e30c708e', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b6', 'filename', 'The file to delete', '{{.filename}}'),

-- Parameters for MoveFile
('5dcd8df8-44c3-767b-a8f5-d1cfef8a7cc9', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b7', 'source', 'The source file to move', '{{.source}}'),
('e4b33064-c8a1-7750-9710-f4d5425ad750', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b7', 'destination', 'The destination file location', '{{.destination}}'),

-- Parameters for CopyFile
('59c4cc6d-c43e-7429-803d-73a2e6d17936', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b8', 'source', 'The source file to copy', '{{.source}}'),
('e5ab9cfc-b9b9-74e1-8d2d-c58d7a4f79b8', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b8', 'destination', 'The destination file location', '{{.destination}}');
`
