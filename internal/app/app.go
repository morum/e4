package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/store/memory"
	"github.com/morum/e4/internal/store/postgres"
	sshtransport "github.com/morum/e4/internal/transport/ssh"
	"github.com/morum/e4/internal/tui/theme"
)

type App struct {
	server *sshtransport.Server
	store  *postgres.Store
}

func New(cfg Config) (*App, error) {
	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	databaseURL := strings.TrimSpace(cfg.DatabaseURL)
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("E4_DATABASE_URL"))
	}
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required: set --database-url or E4_DATABASE_URL")
	}

	store, err := postgres.Open(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}
	if err := store.Migrate(context.Background()); err != nil {
		store.Close()
		return nil, err
	}

	repo := memory.NewRoomRepository()
	lobby := service.NewPersistentLobbyService(repo, store, logger)
	if err := lobby.RestoreOpenGames(context.Background()); err != nil {
		store.Close()
		return nil, err
	}

	themes := theme.Builtin()
	defaultTheme, ok := themes.Get(cfg.Theme)
	if !ok {
		store.Close()
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
		store.Close()
		return nil, err
	}

	return &App{server: server, store: store}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.store.Close()
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
