package room

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
	"github.com/morum/e4/internal/tui/theme"
)

// stubRoom is the minimum service.GameRoom implementation needed to construct
// a room.Model in unit tests. Every method that the Model might invoke during
// our tests is made a no-op; anything else panics so we notice accidental
// coupling.
type stubRoom struct {
	id string
}

func (s stubRoom) ID() string                                         { return s.id }
func (s stubRoom) Snapshot() domain.GameSnapshot                      { return domain.GameSnapshot{RoomID: s.id} }
func (s stubRoom) Subscribe() service.RoomSubscription                { return service.RoomSubscription{} }
func (s stubRoom) JoinPlayer(domain.Participant) (domain.Role, error) { return domain.RoleWhite, nil }
func (s stubRoom) AddWatcher(domain.Participant) error                { return nil }
func (s stubRoom) Leave(string) bool                                  { return false }
func (s stubRoom) SubmitMove(string, string) error                    { return nil }
func (s stubRoom) Resign(string) error                                { return nil }

func newReadyRoom(t *testing.T, width, height int) Model {
	t.Helper()
	sub := service.RoomSubscription{Updates: make(chan domain.GameSnapshot), Cancel: func() {}}
	m := New(domain.Participant{ID: "p1", Nickname: "tester"}, domain.RoleWhite, stubRoom{id: "ABC123"}, sub)

	// Size the model, then feed it a fully-populated snapshot so View renders.
	m, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: height})

	snap := domain.GameSnapshot{
		RoomID:        "ABC123",
		Status:        domain.RoomStatusActive,
		Turn:          "white",
		WhiteName:     "alice",
		BlackName:     "bob",
		WhiteTimeLeft: 5 * time.Minute,
		BlackTimeLeft: 5 * time.Minute,
		Board:         domain.BoardState{Squares: map[string]domain.BoardPiece{}},
		Moves:         []string{"e4", "e5", "Nf3", "Nc6"},
	}
	m, _ = m.Update(SnapshotMsg(snap))
	return m
}

func TestRoomViewDoesNotOverflowNarrowTerminal(t *testing.T) {
	tr := theme.Builtin().Default()

	// A deliberately small SSH window — pre-fix this would overflow both the
	// header (single unbroken line) and the clocks (forced to 32 cols).
	m := newReadyRoom(t, 40, 18)
	out := m.View(tr)

	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 40 {
			t.Fatalf("rendered line overflowed width: %d cols in %q", w, line)
		}
	}

	// The room ID must still appear somewhere in the output — if the header
	// got silently truncated to nothing, that's a regression too.
	if !strings.Contains(out, "ABC123") {
		t.Fatalf("expected room ID in output, got:\n%s", out)
	}
}

func TestCtrlCInRoomRoutesThroughLeaveRequestWithQuit(t *testing.T) {
	sub := service.RoomSubscription{Updates: make(chan domain.GameSnapshot), Cancel: func() {}}
	m := New(domain.Participant{ID: "p1", Nickname: "tester"}, domain.RoleWhite, stubRoom{id: "XYZ"}, sub)

	// Blur input so the ctrl+c keybinding path doesn't get swallowed by the
	// text input. The real program sends ctrl+c regardless of focus state.
	m.input.Blur()
	m.inputOn = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected a command from ctrl+c")
	}
	got := cmd()
	req, ok := got.(LeaveRequestMsg)
	if !ok {
		t.Fatalf("expected LeaveRequestMsg, got %T (%v)", got, got)
	}
	if !req.Quit {
		t.Fatal("expected Quit=true so the parent app can tea.Quit after cleanup")
	}
}
