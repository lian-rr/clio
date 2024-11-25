package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lian-rr/clio/command"
	"github.com/lian-rr/clio/command/store"
	"github.com/lian-rr/clio/command/teacher"
	"github.com/lian-rr/clio/command/teacher/openai"
	"github.com/lian-rr/clio/config"
	"github.com/lian-rr/clio/tui"
)

const (
	debugLogger  = "CLIO_DEBUG"
	storePathEnv = "CLIO_STORE_PATH"
)

func main() {
	// exit once
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var storePath string
	if path, ok := os.LookupEnv(storePathEnv); ok {
		storePath = path
	}

	logger, cancel, err := initLogger()
	if err != nil {
		return err
	}
	defer func() {
		_ = cancel()
	}()

	cfg, err := config.New(ctx, storePath, logger)
	if err != nil {
		return err
	}

	sqlStore, err := store.NewSql(logger, store.WithSqliteDriver(ctx, cfg.BasePath))
	if err != nil {
		slog.Error("error initializing the local store", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := sqlStore.Close(); err != nil {
			logger.Warn("error closing store", slog.Any("error", err))
			return
		}
		logger.Debug("Store closed successfully")
	}()

	manager, err := command.NewManager(sqlStore)
	if err != nil {
		slog.Error("error starting command manager", slog.Any("error", err))
		return err
	}

	var teach *teacher.Teacher
	if tch, ok := newTeacher(cfg.Teacher, logger); ok {
		teach = &tch
	}

	if teach != nil {
		logger.Info("teacher loaded successfully", slog.String("teacher type", string(cfg.Teacher.Type)))
	}

	ui, err := tui.New(ctx, &manager, logger, teach)
	if err != nil {
		return err
	}

	if err := ui.Start(); err != nil {
		return err
	}

	return nil
}

func initLogger() (logger *slog.Logger, close func() error, err error) {
	logLevel := slog.LevelInfo
	if _, ok := os.LookupEnv(debugLogger); ok {
		logLevel = slog.LevelDebug
	}

	file, err := os.OpenFile("/tmp/clio.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, nil, err
	}

	logger = slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	return logger, file.Close, nil
}

func newTeacher(cfg config.TeacherConfig, logger *slog.Logger) (teacher.Teacher, bool) {
	if !cfg.Enabled {
		return teacher.Teacher{}, false
	}

	var source teacher.Source
	switch cfg.Type {
	default:
		opts := make([]openai.OptFunc, 0)
		if cfg.OpenAI.Url != "" {
			opts = append(opts, openai.WithBaseUrl(cfg.OpenAI.Url))
		}
		if cfg.OpenAI.Model != "" {
			opts = append(opts, openai.WithModel(cfg.OpenAI.Model))
		}
		if cfg.OpenAI.CustomPrompt != "" {
			opts = append(opts, openai.WithCustomContext(cfg.OpenAI.ApiKey))
		}

		source = openai.New(logger, cfg.OpenAI.ApiKey, opts...)
	}

	return teacher.New(source, logger), true
}
