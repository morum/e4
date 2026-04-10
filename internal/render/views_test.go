package render

import (
	"strings"
	"testing"
	"time"

	"chessh/internal/domain"
)

func TestPromptUsesContextualModes(t *testing.T) {
	ctx := Context{Width: 100, ANSI: false, Role: domain.RoleWhite, Orientation: OrientationWhite}
	active := &domain.GameSnapshot{RoomID: "ROOM01", Status: domain.RoomStatusActive, Turn: "white"}
	watching := &domain.GameSnapshot{RoomID: "ROOM02", Status: domain.RoomStatusActive, Turn: "black"}
	waiting := &domain.GameSnapshot{RoomID: "ROOM03", Status: domain.RoomStatusWaiting}

	if got := Prompt(ctx, nil, domain.RoleNone); got != "lobby> " {
		t.Fatalf("expected lobby prompt, got %q", got)
	}
	if got := Prompt(ctx, active, domain.RoleWhite); got != "white move> " {
		t.Fatalf("expected white move prompt, got %q", got)
	}
	if got := Prompt(ctx, watching, domain.RoleWatcher); got != "watch[ROOM02]> " {
		t.Fatalf("expected watcher prompt, got %q", got)
	}
	if got := Prompt(ctx, waiting, domain.RoleBlack); got != "room[ROOM03]> " {
		t.Fatalf("expected waiting room prompt, got %q", got)
	}
}

func TestRenderBoardFlipsForBlackOrientation(t *testing.T) {
	board := domain.BoardState{
		Squares: map[string]domain.BoardPiece{
			"a8": {Color: "black", Symbol: "r"},
			"h1": {Color: "white", Symbol: "R"},
		},
	}

	whiteLines := renderBoard(Context{Width: 120, ANSI: false, Orientation: OrientationWhite}, board)
	blackLines := renderBoard(Context{Width: 120, ANSI: false, Orientation: OrientationBlack}, board)

	if !strings.Contains(whiteLines[0], "a  b  c  d  e  f  g  h") {
		t.Fatalf("expected white orientation file order, got %q", whiteLines[0])
	}
	if !strings.Contains(blackLines[0], "h  g  f  e  d  c  b  a") {
		t.Fatalf("expected black orientation file order, got %q", blackLines[0])
	}
	if !strings.Contains(whiteLines[1], " r ") {
		t.Fatalf("expected black rook on top row for white orientation, got %q", whiteLines[1])
	}
	if !strings.Contains(blackLines[1], " R ") {
		t.Fatalf("expected white rook on top row for black orientation, got %q", blackLines[1])
	}
}

func TestRoomViewUsesAnsiAndStatusFooter(t *testing.T) {
	snapshot := sampleSnapshot()
	ctx := Context{
		Width:       120,
		ANSI:        true,
		Role:        domain.RoleWhite,
		Orientation: OrientationWhite,
		Status:      StatusLine{Kind: StatusSuccess, Message: "Created room ROOM01 as White."},
	}

	view := RoomView(ctx, snapshot, "alice", domain.RoleWhite)
	if !strings.Contains(view, "\x1b[") {
		t.Fatalf("expected ANSI styling in room view")
	}
	if !strings.Contains(view, "ROOM ROOM01") {
		t.Fatalf("expected room header in view")
	}
	if !strings.Contains(view, "Created room ROOM01 as White.") {
		t.Fatalf("expected status footer to include transient status")
	}
	if !strings.Contains(view, "Recent Moves") {
		t.Fatalf("expected moves panel in room view")
	}
}

func TestLobbyViewGroupsRoomSections(t *testing.T) {
	ctx := Context{Width: 120, ANSI: false}
	rooms := []domain.RoomSummary{
		{ID: "OPEN01", Status: domain.RoomStatusWaiting, TimeControl: mustTimeControl("3|2"), WhiteName: "alice", HasOpenSeat: true},
		{ID: "LIVE01", Status: domain.RoomStatusActive, TimeControl: mustTimeControl("10|0"), WhiteName: "alice", BlackName: "bob", Turn: "black"},
		{ID: "DONE01", Status: domain.RoomStatusFinished, TimeControl: mustTimeControl("5|0"), WhiteName: "alice", BlackName: "bob", Outcome: "1-0", Method: "checkmate"},
	}

	view := LobbyView(ctx, "alice", rooms)
	for _, want := range []string{"Open Rooms", "Active Games", "Finished Games", "OPEN01", "LIVE01", "DONE01"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected lobby view to contain %q", want)
		}
	}
}

func sampleSnapshot() domain.GameSnapshot {
	return domain.GameSnapshot{
		RoomID:       "ROOM01",
		Status:       domain.RoomStatusActive,
		TimeControl:  mustTimeControl("3|2"),
		WhiteName:    "alice",
		BlackName:    "bob",
		WatcherCount: 2,
		Turn:         "white",
		Board: domain.BoardState{
			Squares: map[string]domain.BoardPiece{
				"e4": {Color: "white", Symbol: "P"},
				"e5": {Color: "black", Symbol: "p"},
				"e1": {Color: "white", Symbol: "K"},
				"e8": {Color: "black", Symbol: "k"},
			},
			LastMoveFrom: "e2",
			LastMoveTo:   "e4",
		},
		Moves:         []string{"e4", "e5", "Nf3"},
		WhiteTimeLeft: 2*time.Minute + 31*time.Second,
		BlackTimeLeft: 18 * time.Second,
		LastEvent:     "alice played Nf3.",
	}
}

func mustTimeControl(raw string) domain.TimeControl {
	tc, err := domain.ParseTimeControl(raw)
	if err != nil {
		panic(err)
	}
	return tc
}
