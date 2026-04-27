package service

import (
	"testing"
	"time"

	"github.com/morum/e4/internal/clock"
	"github.com/morum/e4/internal/domain"

	"github.com/notnil/chess"
)

func TestRoomLifecycleAndSANMoves(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|2")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	room := NewRoom("TEST01", tc, nil)
	white := domain.Participant{ID: "p1", Nickname: "alice"}
	black := domain.Participant{ID: "p2", Nickname: "bob"}

	role, err := room.JoinPlayer(white)
	if err != nil {
		t.Fatalf("JoinPlayer(white) returned error: %v", err)
	}
	if role != domain.RoleWhite {
		t.Fatalf("expected white role, got %s", role)
	}

	role, err = room.JoinPlayer(black)
	if err != nil {
		t.Fatalf("JoinPlayer(black) returned error: %v", err)
	}
	if role != domain.RoleBlack {
		t.Fatalf("expected black role, got %s", role)
	}

	if err := room.SubmitMove(white.ID, "e4"); err != nil {
		t.Fatalf("SubmitMove returned error: %v", err)
	}

	snapshot := room.Snapshot()
	if snapshot.Status != domain.RoomStatusActive {
		t.Fatalf("expected active room, got %s", snapshot.Status)
	}
	if len(snapshot.Moves) != 1 || snapshot.Moves[0] != "e4" {
		t.Fatalf("expected SAN move history to contain e4, got %#v", snapshot.Moves)
	}
	if snapshot.Turn != "black" {
		t.Fatalf("expected black to move next, got %q", snapshot.Turn)
	}
	if snapshot.Board.LastMoveFrom != "e2" || snapshot.Board.LastMoveTo != "e4" {
		t.Fatalf("expected last move squares to be tracked, got %q -> %q", snapshot.Board.LastMoveFrom, snapshot.Board.LastMoveTo)
	}
	if piece := snapshot.Board.Squares["e4"]; piece.Symbol != "♙" {
		t.Fatalf("expected board snapshot to use piece glyphs, got %q", piece.Symbol)
	}

	if err := room.Resign(black.ID); err != nil {
		t.Fatalf("Resign returned error: %v", err)
	}

	finished := room.Snapshot()
	if finished.Status != domain.RoomStatusFinished {
		t.Fatalf("expected finished room, got %s", finished.Status)
	}
	if finished.Outcome == "" {
		t.Fatalf("expected outcome to be set after resignation")
	}

	if room.Leave(white.ID) {
		t.Fatalf("expected room to remain until final participant leaves")
	}
	if !room.Leave(black.ID) {
		t.Fatalf("expected room to be empty after both players leave")
	}
}

func TestRoomSnapshotTracksCheckSquare(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|2")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	room := NewRoom("TEST02", tc, nil)
	white := domain.Participant{ID: "p1", Nickname: "alice"}
	black := domain.Participant{ID: "p2", Nickname: "bob"}

	if _, err := room.JoinPlayer(white); err != nil {
		t.Fatalf("JoinPlayer(white) returned error: %v", err)
	}
	if _, err := room.JoinPlayer(black); err != nil {
		t.Fatalf("JoinPlayer(black) returned error: %v", err)
	}

	moves := []struct {
		playerID string
		move     string
	}{
		{white.ID, "e4"},
		{black.ID, "e5"},
		{white.ID, "Qh5"},
		{black.ID, "Nc6"},
		{white.ID, "Bc4"},
		{black.ID, "Nf6"},
		{white.ID, "Qxf7"},
	}

	for _, item := range moves {
		if err := room.SubmitMove(item.playerID, item.move); err != nil {
			t.Fatalf("SubmitMove(%s) returned error: %v", item.move, err)
		}
	}

	snapshot := room.Snapshot()
	if snapshot.Board.CheckSquare != "e8" {
		t.Fatalf("expected black king on e8 to be highlighted in check, got %q", snapshot.Board.CheckSquare)
	}
	if snapshot.Board.LastMoveFrom != "h5" || snapshot.Board.LastMoveTo != "f7" {
		t.Fatalf("expected last move to be tracked as h5 -> f7, got %q -> %q", snapshot.Board.LastMoveFrom, snapshot.Board.LastMoveTo)
	}
}

func TestActiveRoomBroadcastsClockTicks(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|0")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	room := NewRoom("TEST03", tc, nil)
	white := domain.Participant{ID: "p1", Nickname: "alice"}
	black := domain.Participant{ID: "p2", Nickname: "bob"}

	if _, err := room.JoinPlayer(white); err != nil {
		t.Fatalf("JoinPlayer(white) returned error: %v", err)
	}
	if _, err := room.JoinPlayer(black); err != nil {
		t.Fatalf("JoinPlayer(black) returned error: %v", err)
	}

	sub := room.Subscribe()
	defer sub.Cancel()

	initial := <-sub.Updates
	select {
	case next := <-sub.Updates:
		if next.WhiteTimeLeft >= initial.WhiteTimeLeft {
			t.Fatalf("expected white clock to tick down, got %v then %v", initial.WhiteTimeLeft, next.WhiteTimeLeft)
		}
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("expected active room to broadcast a clock tick")
	}
}

func TestRestoreActiveRoomPausesUntilBothPlayersReconnect(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|0")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	white := domain.Participant{ID: "11111111-1111-1111-1111-111111111111", Nickname: "alice"}
	black := domain.Participant{ID: "22222222-2222-2222-2222-222222222222", Nickname: "bob"}
	room, err := RestoreRoom(PersistedRoom{
		ID:          "REST01",
		Status:      domain.RoomStatusActive,
		TimeControl: tc,
		White:       &white,
		Black:       &black,
		Moves:       []string{"e4", "e5"},
		Clock: clock.Snapshot{
			WhiteRemaining: 3 * time.Second,
			BlackRemaining: 3 * time.Second,
			Increment:      tc.Increment,
		},
	}, nil, nil)
	if err != nil {
		t.Fatalf("RestoreRoom returned error: %v", err)
	}

	before := room.Snapshot()
	if _, err := room.JoinPlayer(white); err != nil {
		t.Fatalf("JoinPlayer(white) returned error: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	afterOne := room.Snapshot()
	if afterOne.WhiteTimeLeft != before.WhiteTimeLeft {
		t.Fatalf("expected restored game to remain paused until both players reconnect, got %v then %v", before.WhiteTimeLeft, afterOne.WhiteTimeLeft)
	}

	if _, err := room.JoinPlayer(black); err != nil {
		t.Fatalf("JoinPlayer(black) returned error: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	afterBoth := room.Snapshot()
	if afterBoth.WhiteTimeLeft >= afterOne.WhiteTimeLeft {
		t.Fatalf("expected side-to-move clock to run after both reconnect, got %v then %v", afterOne.WhiteTimeLeft, afterBoth.WhiteTimeLeft)
	}
}

func TestClosedRoomMethodsReturnWithoutDeadlock(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|0")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	room := NewRoom("TEST04", tc, nil)
	white := domain.Participant{ID: "p1", Nickname: "alice"}
	if _, err := room.JoinPlayer(white); err != nil {
		t.Fatalf("JoinPlayer returned error: %v", err)
	}
	if !room.Leave(white.ID) {
		t.Fatal("expected room to be empty after last participant leaves")
	}
	<-room.done

	assertReturns(t, "Snapshot", func() {
		snap := room.Snapshot()
		if snap.RoomID != "TEST04" || snap.Status != domain.RoomStatusFinished {
			t.Fatalf("unexpected closed snapshot: %#v", snap)
		}
	})
	assertReturns(t, "Subscribe", func() {
		sub := room.Subscribe()
		if _, ok := <-sub.Updates; ok {
			t.Fatal("expected closed subscription channel")
		}
		sub.Cancel()
	})
	assertReturns(t, "JoinPlayer", func() {
		if _, err := room.JoinPlayer(domain.Participant{ID: "p2", Nickname: "bob"}); err != ErrRoomClosed {
			t.Fatalf("expected ErrRoomClosed, got %v", err)
		}
	})
	assertReturns(t, "AddWatcher", func() {
		if err := room.AddWatcher(domain.Participant{ID: "w1", Nickname: "watcher"}); err != ErrRoomClosed {
			t.Fatalf("expected ErrRoomClosed, got %v", err)
		}
	})
	assertReturns(t, "Leave", func() {
		if !room.Leave("p2") {
			t.Fatal("expected Leave on a closed room to report empty")
		}
	})
	assertReturns(t, "SubmitMove", func() {
		if err := room.SubmitMove("p1", "e4"); err != ErrRoomClosed {
			t.Fatalf("expected ErrRoomClosed, got %v", err)
		}
	})
	assertReturns(t, "Resign", func() {
		if err := room.Resign("p1"); err != ErrRoomClosed {
			t.Fatalf("expected ErrRoomClosed, got %v", err)
		}
	})
}

func TestBroadcastDropsWhenSubscriberCannotReceive(t *testing.T) {
	tc, err := domain.ParseTimeControl("3|0")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}
	state := roomState{
		timeControl: tc,
		status:      domain.RoomStatusWaiting,
		game:        chess.NewGame(),
		clock:       clock.New(tc),
		watchers:    make(map[string]domain.Participant),
		subs: map[int]chan domain.GameSnapshot{
			1: make(chan domain.GameSnapshot),
		},
	}

	assertReturns(t, "broadcast", func() {
		state.broadcast("TEST05")
	})
}

func assertReturns(t *testing.T, name string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("%s did not return", name)
	}
}
