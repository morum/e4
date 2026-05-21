package tui

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/app"
	"github.com/morum/e4/internal/tui/theme"
	gossh "golang.org/x/crypto/ssh"
)

// nonPTYMessage is what we tell clients that attach without an interactive
// terminal (exec, sftp, etc.). Kept as a package-level const so tests can
// assert on it.
const nonPTYMessage = "e4 requires an interactive terminal. Try: ssh -t -p 2222 <host>"

var entropySource io.Reader = rand.Reader
var fallbackIDSeq atomic.Uint64

// rejectNonPTY writes the no-tty message and exits the session. Split out
// from Handler so it can be unit-tested without constructing a full Session.
func rejectNonPTY(w io.Writer, exit func(int) error) {
	fmt.Fprintln(w, nonPTYMessage)
	if exit != nil {
		_ = exit(1)
	}
}

// Handler returns a bubbletea.Handler that serves the e4 TUI over SSH.
// Each session gets a fresh app.Model wired to the shared lobby service
// and theme registry.
func Handler(lobby *service.LobbyService, registry *theme.Registry, defaultTheme theme.Theme, logger *slog.Logger) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, active := sess.Pty()
		if !active {
			// activeterm.Middleware should catch this before we get here, but
			// guard against future middleware-ordering changes: tell the user
			// explicitly that an interactive TTY is required rather than
			// returning (nil, nil) and letting bubbletea fail silently.
			rejectNonPTY(sess.Stderr(), sess.Exit)
			return nil, nil
		}

		sessionID := randomID()
		key := sess.PublicKey()
		if key == nil {
			fmt.Fprintln(sess.Stderr(), "e4 requires SSH public key authentication.")
			_ = sess.Exit(1)
			return nil, nil
		}
		participant := domain.Participant{
			ID:               sessionID,
			SessionID:        sessionID,
			KeyFingerprint:   gossh.FingerprintSHA256(key),
			SSHAuthorizedKey: strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key))),
			SSHKeyType:       key.Type(),
			Nickname:         guestNickname(sessionID),
		}

		selected := defaultTheme
		if !supportsColorTheme(pty.Term) {
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
		go func() {
			<-sess.Context().Done()
			model.CleanupSession()
		}()
		return model, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	}
}

func randomID() string {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(entropySource, buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fallbackID()
}

func fallbackID() string {
	buf := make([]byte, 8)
	seed := uint64(time.Now().UnixNano()) ^ fallbackIDSeq.Add(1)
	binary.BigEndian.PutUint64(buf, seed)
	return hex.EncodeToString(buf)
}

func guestNickname(id string) string {
	if len(id) < 4 {
		return "guest-" + id
	}
	return "guest-" + id[:4]
}

func supportsColorTheme(term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" || term == "dumb" {
		return false
	}
	return strings.Contains(term, "256color") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(term, "24bit") ||
		strings.Contains(term, "direct") ||
		strings.HasPrefix(term, "xterm-kitty") ||
		strings.HasPrefix(term, "alacritty") ||
		strings.HasPrefix(term, "foot") ||
		term == "wezterm"
}
