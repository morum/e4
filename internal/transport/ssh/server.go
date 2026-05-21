package ssh

import (
	"context"
	"log/slog"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui"
	"github.com/morum/e4/internal/tui/theme"
)

type Config struct {
	ListenAddr   string
	HostKeyPath  string
	Logger       *slog.Logger
	Themes       *theme.Registry
	DefaultTheme theme.Theme
}

type Server struct {
	server *ssh.Server
	logger *slog.Logger
}

func NewServer(cfg Config, lobby *service.LobbyService) (*Server, error) {
	handler := tui.Handler(lobby, cfg.Themes, cfg.DefaultTheme, cfg.Logger)

	srv, err := wish.NewServer(
		wish.WithAddress(cfg.ListenAddr),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithIdleTimeout(8*time.Hour),
		wish.WithMaxTimeout(24*time.Hour),
		wish.WithBanner("Welcome to e4\n"),
		wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			return key != nil
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(handler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Server{server: srv, logger: cfg.Logger}, nil
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	s.logInfo("ssh server starting", "listen_addr", s.server.Addr)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == nil || err == ssh.ErrServerClosed {
			return nil
		}
		s.logError("ssh server stopped with error", "error", err)
		return err
	case <-ctx.Done():
		s.logInfo("ssh server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
		err := <-errCh
		if err == nil || err == ssh.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) logInfo(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Info(msg, attrs...)
}

func (s *Server) logError(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Error(msg, attrs...)
}
