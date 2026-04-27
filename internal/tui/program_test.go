package tui

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestRejectNonPTYWritesHelpAndExitsNonZero(t *testing.T) {
	var buf bytes.Buffer
	exited := -1
	rejectNonPTY(&buf, func(code int) error {
		exited = code
		return nil
	})

	out := buf.String()
	if !strings.Contains(out, "interactive terminal") {
		t.Fatalf("expected message to mention interactive terminal, got %q", out)
	}
	if exited != 1 {
		t.Fatalf("expected exit code 1, got %d", exited)
	}
}

func TestRejectNonPTYToleratesNilExit(t *testing.T) {
	var buf bytes.Buffer
	// Must not panic when Exit is unavailable (defensive — Handler always
	// supplies sess.Exit, but keeping the helper tolerant keeps it easy to
	// reuse from other call sites).
	rejectNonPTY(&buf, nil)
	if buf.Len() == 0 {
		t.Fatal("expected message to be written even when exit is nil")
	}
}

func TestGuestNicknameUsesPrefix(t *testing.T) {
	if got := guestNickname("deadbeefcafe"); got != "guest-dead" {
		t.Fatalf("expected guest-dead, got %q", got)
	}
	if got := guestNickname("ab"); got != "guest-ab" {
		t.Fatalf("expected guest-ab for short IDs, got %q", got)
	}
}

func TestRandomIDFallbackRemainsUnique(t *testing.T) {
	prev := entropySource
	entropySource = failingReader{}
	fallbackIDSeq.Store(0)
	defer func() {
		entropySource = prev
		fallbackIDSeq.Store(0)
	}()

	first := randomID()
	second := randomID()

	if len(first) != 16 || len(second) != 16 {
		t.Fatalf("expected 16 hex chars from fallback IDs, got %q and %q", first, second)
	}
	if first == second {
		t.Fatalf("expected fallback IDs to remain unique, got %q twice", first)
	}
}

func TestSupportsColorThemeRequiresColorCapableTerminals(t *testing.T) {
	cases := map[string]bool{
		"":                  false,
		"dumb":              false,
		"vt100":             false,
		"linux":             false,
		"ansi":              false,
		"xterm":             false,
		"screen":            false,
		"xterm-256color":    true,
		"screen-256color":   true,
		"tmux-24bit":        true,
		"xterm-direct":      true,
		"xterm-kitty":       true,
		"alacritty":         true,
		"foot":              true,
		"wezterm":           true,
		"wezterm-truecolor": true,
	}
	for term, want := range cases {
		if got := supportsColorTheme(term); got != want {
			t.Errorf("supportsColorTheme(%q) = %v, want %v", term, got, want)
		}
	}
}
