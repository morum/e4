package service

import (
	"testing"

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
	if snapshot.Turn != "b" && snapshot.Turn != "black" {
		t.Fatalf("expected black to move next, got %q", snapshot.Turn)
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
