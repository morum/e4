package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/store/memory"
	sshtransport "github.com/morum/e4/internal/transport/ssh"
	"github.com/morum/e4/internal/tui/theme"
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

	themes := theme.Builtin()
	defaultTheme, ok := themes.Get(cfg.Theme)
	if !ok {
		return nil, fmt.Errorf("unknown theme %q (available: %v)", cfg.Theme, themes.Names())
	}

	server, err := sshtransport.NewServer(sshtransport.Config{
		ListenAddr:   cfg.ListenAddr,
		HostKeyPath:  cfg.HostKeyPath,
		Logger:       logger,
		Themes:       themes,
		DefaultTheme: defaultTheme,
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
