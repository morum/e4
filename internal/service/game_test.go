package service

import (
	"testing"
	"time"

	"chessh/internal/domain"
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
