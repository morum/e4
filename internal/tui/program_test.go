package tui

import (
	"bytes"
	"strings"
	"testing"
)

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

func TestIsANSICapableHandlesDumbTerminals(t *testing.T) {
	cases := map[string]bool{
		"":       false,
		"dumb":   false,
		"xterm":  true,
		"screen": true,
	}
	for term, want := range cases {
		if got := isANSICapable(term); got != want {
			t.Errorf("isANSICapable(%q) = %v, want %v", term, got, want)
		}
	}
}
