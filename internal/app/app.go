package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"chessh/internal/service"
	"chessh/internal/store/memory"
	sshtransport "chessh/internal/transport/ssh"
)

type App struct {
	server *sshtransport.Server
}

func New(cfg Config) (*App, error) {
	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	repo := memory.NewRoomRepository()
	lobby := service.NewLobbyService(repo, logger)

	server, err := sshtransport.NewServer(sshtransport.Config{
		ListenAddr:  cfg.ListenAddr,
		HostKeyPath: cfg.HostKeyPath,
		Logger:      logger,
	}, lobby)
	if err != nil {
		return nil, err
	}

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.server.ListenAndServe(ctx)
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch raw {
	case "debug":
		return slog.LevelDebug, nil
	case "", "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q", raw)
	}
}
