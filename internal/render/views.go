package render

import (
	"fmt"
	"strings"
	"time"

	"chessh/internal/domain"
)

func LobbyView(ctx Context, nickname string, rooms []domain.RoomSummary) string {
	t := newTheme(ctx.ANSI)
	var b strings.Builder

	b.WriteString(t.title("chessh"))
	b.WriteString("  ")
	b.WriteString(t.section("LOBBY"))
	b.WriteString("\n")
	b.WriteString(t.dim(strings.Repeat("=", 72)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Connected as %s\n", t.accent(nickname)))
	b.WriteString(fmt.Sprintf("Quick start: %s\n\n", t.muted("create 10|0  |  join <room>  |  watch <room>")))

	openRooms, activeRooms, finishedRooms := splitRooms(rooms)
	b.WriteString(renderLobbySection(ctx, "Open Rooms", "Joinable seats ready now.", openRooms))
	b.WriteString("\n")
	b.WriteString(renderLobbySection(ctx, "Active Games", "Live games in progress.", activeRooms))
	b.WriteString("\n")
	b.WriteString(renderLobbySection(ctx, "Finished Games", "Recent completed games.", finishedRooms))
	b.WriteString("\n")
	b.WriteString(renderFooter(ctx, lobbyStatusMessage(rooms), "Commands: list, create <tc>, join <id>, watch <id>, help, quit"))

	return strings.TrimRight(b.String(), "\n")
}

func RoomView(ctx Context, snapshot domain.GameSnapshot, nickname string, role domain.Role) string {
	t := newTheme(ctx.ANSI)
	var b strings.Builder

	b.WriteString(t.title("chessh"))
	b.WriteString("  ")
	b.WriteString(t.section("ROOM " + snapshot.RoomID))
	b.WriteString("  ")
	b.WriteString(t.badge(snapshot.Status))
	b.WriteString("\n")
	b.WriteString(t.dim(strings.Repeat("=", 72)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("You: %s (%s)  Time: %s\n",
		t.accent(nickname),
		t.role(role, roleLabel(role)),
		t.accent(snapshot.TimeControl.String()),
	))
	b.WriteString(renderRoomStatus(ctx, snapshot))
	b.WriteString("\n")

	boardLines := renderBoard(ctx, snapshot.Board)
	movesLines := renderMovesPanel(ctx, snapshot)
	if ctx.Layout() == LayoutWide {
		b.WriteString(joinColumns(boardLines, movesLines, 34))
	} else {
		b.WriteString(strings.Join(boardLines, "\n"))
		b.WriteString("\n\n")
		b.WriteString(strings.Join(movesLines, "\n"))
	}
	b.WriteString("\n\n")
	b.WriteString(renderFooter(ctx, roomStatusMessage(snapshot), roomHint(snapshot, role)))

	return strings.TrimRight(b.String(), "\n")
}

func HelpText(inRoom bool) string {
	if inRoom {
		return "Room commands: help, board, leave, resign, quit, or enter a SAN move like e4 or O-O."
	}
	return "Lobby commands: list, create <tc>, join <id>, watch <id>, help, quit."
}

func Prompt(_ Context, snapshot *domain.GameSnapshot, role domain.Role) string {
	if snapshot == nil {
		return "lobby> "
	}

	if role == domain.RoleWatcher {
		return fmt.Sprintf("watch[%s]> ", snapshot.RoomID)
	}

	if snapshot.Status == domain.RoomStatusActive {
		if role == domain.RoleWhite && snapshot.Turn == "white" {
			return "white move> "
		}
		if role == domain.RoleBlack && snapshot.Turn == "black" {
			return "black move> "
		}
	}

	return fmt.Sprintf("room[%s]> ", snapshot.RoomID)
}

func renderLobbySection(ctx Context, title, subtitle string, rooms []domain.RoomSummary) string {
	t := newTheme(ctx.ANSI)
	var b strings.Builder

	b.WriteString(t.section(title))
	b.WriteString("\n")
	b.WriteString(t.dim(subtitle))
	b.WriteString("\n")

	if len(rooms) == 0 {
		b.WriteString(t.dim("  none\n"))
		return b.String()
	}

	if ctx.Layout() == LayoutWide {
		b.WriteString(t.muted("  ID      TC    Seat/State         Players                    Watch\n"))
		for _, room := range rooms {
			b.WriteString(renderLobbyWideRow(t, room))
			b.WriteString("\n")
		}
		return b.String()
	}

	for _, room := range rooms {
		b.WriteString(renderLobbyCompactRow(t, room))
		b.WriteString("\n")
	}

	return b.String()
}

func renderLobbyWideRow(t theme, room domain.RoomSummary) string {
	seatState := roomStateLabel(room)
	players := fmt.Sprintf("%s vs %s", fallback(room.WhiteName, "open"), fallback(room.BlackName, "open"))
	return fmt.Sprintf("  %-6s  %-4s  %-17s  %-25s  %d",
		t.accent(room.ID),
		room.TimeControl.String(),
		seatState,
		players,
		room.WatcherCount,
	)
}

func renderLobbyCompactRow(t theme, room domain.RoomSummary) string {
	parts := []string{
		fmt.Sprintf("  %s %s", t.accent(room.ID), room.TimeControl.String()),
		fmt.Sprintf("  %s", roomStateLabel(room)),
		fmt.Sprintf("  %s vs %s", fallback(room.WhiteName, "open"), fallback(room.BlackName, "open")),
	}

	if room.WatcherCount > 0 {
		parts = append(parts, fmt.Sprintf("  watchers: %d", room.WatcherCount))
	}

	return strings.Join(parts, "\n")
}

func renderRoomStatus(ctx Context, snapshot domain.GameSnapshot) string {
	t := newTheme(ctx.ANSI)
	whiteActive := snapshot.Status == domain.RoomStatusActive && snapshot.Turn == "white"
	blackActive := snapshot.Status == domain.RoomStatusActive && snapshot.Turn == "black"

	statusText := waitingText(snapshot)
	if snapshot.Outcome != "" {
		statusText = fmt.Sprintf("Result: %s by %s", snapshot.Outcome, snapshot.Method)
	} else if snapshot.Status == domain.RoomStatusActive {
		statusText = fmt.Sprintf("Turn: %s to move", snapshot.Turn)
	}

	return strings.Join([]string{
		fmt.Sprintf("White: %-14s %s", t.playerName("white", fallback(snapshot.WhiteName, "open")), t.clock(formatDuration(snapshot.WhiteTimeLeft), whiteActive, snapshot.WhiteTimeLeft)),
		fmt.Sprintf("Black: %-14s %s", t.playerName("black", fallback(snapshot.BlackName, "open")), t.clock(formatDuration(snapshot.BlackTimeLeft), blackActive, snapshot.BlackTimeLeft)),
		fmt.Sprintf("Watchers: %d  %s", snapshot.WatcherCount, t.accent(statusText)),
	}, "\n")
}

func renderMovesPanel(ctx Context, snapshot domain.GameSnapshot) []string {
	t := newTheme(ctx.ANSI)
	lines := []string{t.section("Recent Moves")}
	if len(snapshot.Moves) == 0 {
		lines = append(lines, t.dim("  none yet"))
		return lines
	}

	start := 0
	maxMoves := 14
	if ctx.Layout() != LayoutWide {
		maxMoves = 10
	}
	if len(snapshot.Moves) > maxMoves {
		start = len(snapshot.Moves) - maxMoves
		if start%2 != 0 {
			start--
		}
	}

	for i := start; i < len(snapshot.Moves); i += 2 {
		moveNo := i/2 + 1
		white := snapshot.Moves[i]
		black := ""
		if i+1 < len(snapshot.Moves) {
			black = snapshot.Moves[i+1]
		}
		if i == len(snapshot.Moves)-1 || i+1 == len(snapshot.Moves)-1 {
			if black != "" {
				black = t.accent(black)
			} else {
				white = t.accent(white)
			}
		}
		lines = append(lines, fmt.Sprintf("  %2d. %-9s %s", moveNo, white, black))
	}

	return lines
}

func renderFooter(ctx Context, fallbackStatus, hint string) string {
	t := newTheme(ctx.ANSI)
	status := fallbackStatus
	statusKind := StatusInfo
	if strings.TrimSpace(ctx.Status.Message) != "" {
		status = ctx.Status.Message
		statusKind = ctx.Status.Kind
	}

	var lines []string
	if strings.TrimSpace(status) != "" {
		lines = append(lines, t.status(statusKind, status))
	}
	if strings.TrimSpace(hint) != "" {
		lines = append(lines, t.muted(hint))
	}
	return strings.Join(lines, "\n")
}

func roomStatusMessage(snapshot domain.GameSnapshot) string {
	if snapshot.LastEvent != "" {
		return snapshot.LastEvent
	}
	if snapshot.Status == domain.RoomStatusWaiting {
		return waitingText(snapshot)
	}
	if snapshot.Outcome != "" {
		return fmt.Sprintf("Game finished: %s by %s.", snapshot.Outcome, snapshot.Method)
	}
	return ""
}

func lobbyStatusMessage(rooms []domain.RoomSummary) string {
	if len(rooms) == 0 {
		return "No rooms yet. Create one with `create 10|0`."
	}

	openRooms, activeRooms, finishedRooms := splitRooms(rooms)
	return fmt.Sprintf("%d open, %d active, %d finished.", len(openRooms), len(activeRooms), len(finishedRooms))
}

func roomHint(snapshot domain.GameSnapshot, role domain.Role) string {
	if role == domain.RoleWatcher {
		return "Commands: board, leave, quit. You receive live updates automatically."
	}
	if snapshot.Status == domain.RoomStatusWaiting {
		return "Commands: leave, quit. Waiting for the other seat to fill."
	}
	return "Commands: <SAN move>, board, resign, leave, quit. Examples: e4, Nf3, O-O."
}

func splitRooms(rooms []domain.RoomSummary) ([]domain.RoomSummary, []domain.RoomSummary, []domain.RoomSummary) {
	openRooms := make([]domain.RoomSummary, 0)
	activeRooms := make([]domain.RoomSummary, 0)
	finishedRooms := make([]domain.RoomSummary, 0)
	for _, room := range rooms {
		switch {
		case room.Status == domain.RoomStatusActive:
			activeRooms = append(activeRooms, room)
		case room.Status == domain.RoomStatusWaiting && room.HasOpenSeat:
			openRooms = append(openRooms, room)
		default:
			finishedRooms = append(finishedRooms, room)
		}
	}
	return openRooms, activeRooms, finishedRooms
}

func roomStateLabel(room domain.RoomSummary) string {
	if room.Status == domain.RoomStatusFinished {
		return fmt.Sprintf("finished %s %s", room.Outcome, room.Method)
	}
	if room.Status == domain.RoomStatusActive {
		return fmt.Sprintf("%s to move", room.Turn)
	}
	if room.WhiteName == "" {
		return "white seat open"
	}
	if room.BlackName == "" {
		return "black seat open"
	}
	return "waiting"
}

func waitingText(snapshot domain.GameSnapshot) string {
	if snapshot.WhiteName == "" {
		return "Waiting for White to join."
	}
	if snapshot.BlackName == "" {
		return "Waiting for Black to join."
	}
	return "Waiting for players."
}

func joinColumns(left, right []string, leftWidth int) string {
	rowCount := len(left)
	if len(right) > rowCount {
		rowCount = len(right)
	}

	var b strings.Builder
	for i := 0; i < rowCount; i++ {
		leftLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}
		b.WriteString(fmt.Sprintf("%-*s  %s\n", leftWidth, leftLine, rightLine))
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Round(time.Second) / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func roleLabel(role domain.Role) string {
	if role == domain.RoleNone {
		return "lobby"
	}
	return string(role)
}
