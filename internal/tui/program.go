package tui

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/app"
	"github.com/morum/e4/internal/tui/theme"
)

// Handler returns a bubbletea.Handler that serves the e4 TUI over SSH.
// Each session gets a fresh app.Model wired to the shared lobby service
// and theme registry.
func Handler(lobby *service.LobbyService, registry *theme.Registry, defaultTheme theme.Theme, logger *slog.Logger) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, active := sess.Pty()
		if !active {
			return nil, nil
		}

		sessionID := randomID()
		participant := domain.Participant{
			ID:       sessionID,
			Nickname: guestNickname(sessionID),
		}

		selected := defaultTheme
		if !isANSICapable(pty.Term) {
			if m, ok := registry.Get("mono"); ok {
				selected = m
			}
		}

		if logger != nil {
			logger.Info("tui session started",
				"session_id", sessionID,
				"remote_addr", sess.RemoteAddr().String(),
				"ssh_user", sess.User(),
				"term", pty.Term,
				"theme", selected.Name,
			)
		}

		model := app.New(participant, lobby, registry, selected)
		return model, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	}
}

func randomID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}

func guestNickname(id string) string {
	if len(id) < 4 {
		return "guest-" + id
	}
	return "guest-" + id[:4]
}

func isANSICapable(term string) bool {
	if term == "" || term == "dumb" {
		return false
	}
	return true
}
