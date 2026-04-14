package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/store/memory"
	"github.com/morum/e4/internal/tui/lobby"
	"github.com/morum/e4/internal/tui/room"
	"github.com/morum/e4/internal/tui/theme"
)

func newTestModel(t *testing.T) (Model, *theme.Registry, *service.LobbyService) {
	t.Helper()
	registry := theme.NewRegistry()
	registry.Register(theme.Theme{Name: "A"})
	registry.Register(theme.Theme{Name: "B"})
	registry.Register(theme.Theme{Name: "C"})
	_ = registry.SetDefault("A")

	lobbySvc := service.NewLobbyService(memory.NewRoomRepository(), nil)
	m := New(domain.Participant{ID: "sess-1", Nickname: "tester"}, lobbySvc, registry, registry.Default())
	return m, registry, lobbySvc
}

func TestCycleThemeFromLobby(t *testing.T) {
	m, _, _ := newTestModel(t)
	if got := m.ThemeName(); got != "A" {
		t.Fatalf("expected starting theme A, got %q", got)
	}

	updated, _ := m.Update(lobby.CycleThemeMsg{})
	m = updated.(Model)
	if got := m.ThemeName(); got != "B" {
		t.Fatalf("expected theme B after one cycle, got %q", got)
	}

	updated, _ = m.Update(room.CycleThemeMsg{})
	m = updated.(Model)
	if got := m.ThemeName(); got != "C" {
		t.Fatalf("expected theme C after second cycle, got %q", got)
	}

	updated, _ = m.Update(lobby.CycleThemeMsg{})
	m = updated.(Model)
	if got := m.ThemeName(); got != "A" {
		t.Fatalf("expected theme A after wrap-around, got %q", got)
	}
}

func TestSetThemeByName(t *testing.T) {
	m, _, _ := newTestModel(t)

	updated, _ := m.Update(lobby.SetThemeMsg{Name: "  C  "})
	m = updated.(Model)
	if got := m.ThemeName(); got != "C" {
		t.Fatalf("expected theme C after SetThemeMsg, got %q", got)
	}

	// An unknown theme must leave the current theme untouched.
	updated, _ = m.Update(room.SetThemeMsg{Name: "nope"})
	m = updated.(Model)
	if got := m.ThemeName(); got != "C" {
		t.Fatalf("expected unknown theme to be ignored, got %q", got)
	}
}

func TestLeaveRoomClearsHandleAndReturnsToLobby(t *testing.T) {
	m, _, lobbySvc := newTestModel(t)

	// Boot through the join screen and create a real room so the app holds a
	// live subscription, mirroring the runtime flow.
	m = advancePastJoin(t, m, "tester")
	m = enterRoom(t, m, lobbySvc)

	if !m.InRoom() {
		t.Fatal("expected app to be in a room after enterRoom")
	}

	updated, _ := m.Update(room.LeaveRequestMsg{Reason: "test leave"})
	m = updated.(Model)

	if m.InRoom() {
		t.Fatal("expected room handle to be cleared after LeaveRequestMsg")
	}
	if m.Screen() != ScreenLobby {
		t.Fatalf("expected ScreenLobby after leave, got %v", m.Screen())
	}
}

func TestCtrlCQuitFromRoomTearsDownBeforeQuitting(t *testing.T) {
	m, _, lobbySvc := newTestModel(t)
	m = advancePastJoin(t, m, "tester")
	m = enterRoom(t, m, lobbySvc)

	if !m.InRoom() {
		t.Fatal("expected app to be in a room")
	}

	// The Quit variant of LeaveRequestMsg is what ctrl+c in the room sends;
	// it must still run the cleanup path before asking bubbletea to quit.
	updated, cmd := m.Update(room.LeaveRequestMsg{Reason: "bye", Quit: true})
	m = updated.(Model)

	if m.InRoom() {
		t.Fatal("expected room handle to be cleared even on quit")
	}
	if cmd == nil {
		t.Fatal("expected a non-nil command (tea.Quit) after quit request")
	}
	if got := cmd(); got == nil {
		t.Fatal("expected command to resolve to a quit message")
	} else if _, ok := got.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", got)
	}
}

func advancePastJoin(t *testing.T, m Model, nickname string) Model {
	t.Helper()
	// Send the nickname submission message that the join screen would emit.
	// This path is exercised end-to-end in the real app; we just drive it
	// synchronously here.
	updated, _ := m.Update(nicknameSubmittedMsg{Nickname: nickname})
	return updated.(Model)
}

func enterRoom(t *testing.T, m Model, lobbySvc *service.LobbyService) Model {
	t.Helper()
	tc, err := domain.ParseTimeControl("5|0")
	if err != nil {
		t.Fatalf("failed to parse time control: %v", err)
	}
	gameRoom, role, err := lobbySvc.CreateGame(domain.Participant{ID: "sess-1", Nickname: "tester"}, tc)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}
	sub := gameRoom.Subscribe()

	updated, _ := m.Update(lobby.EnterRoomMsg{Room: gameRoom, Role: role, Sub: sub})
	return updated.(Model)
}
