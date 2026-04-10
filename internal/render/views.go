package render

import (
	"fmt"
	"strings"
	"time"

	"chessh/internal/domain"
)

func LobbyView(nickname string, rooms []domain.RoomSummary) string {
	var b strings.Builder
	b.WriteString("chessh\n")
	b.WriteString("======\n\n")
	fmt.Fprintf(&b, "Connected as %s\n\n", nickname)

	if len(rooms) == 0 {
		b.WriteString("No rooms yet. Create one with `create 10|0`.\n")
	} else {
		b.WriteString("Rooms\n")
		b.WriteString("ID      TC    Status    White        Black        Watchers  Turn/Result\n")
		b.WriteString("----------------------------------------------------------------------\n")
		for _, room := range rooms {
			statusText := string(room.Status)
			turnText := room.Turn
			if room.Outcome != "" {
				turnText = fmt.Sprintf("%s %s", room.Outcome, room.Method)
			}
			fmt.Fprintf(&b, "%-6s  %-4s  %-8s  %-11s  %-11s  %-8d  %s\n",
				room.ID,
				room.TimeControl.String(),
				statusText,
				fallback(room.WhiteName, "open"),
				fallback(room.BlackName, "open"),
				room.WatcherCount,
				turnText,
			)
		}
	}

	b.WriteString("\nCommands\n")
	b.WriteString("  list                refresh the lobby\n")
	b.WriteString("  create <tc>         create a room as White, e.g. create 10|0\n")
	b.WriteString("  join <id>           take the open seat in a room\n")
	b.WriteString("  watch <id>          watch a room\n")
	b.WriteString("  help                show help\n")
	b.WriteString("  quit                disconnect\n")
	return b.String()
}

func RoomView(snapshot domain.GameSnapshot, nickname string, role domain.Role) string {
	var b strings.Builder
	b.WriteString("chessh\n")
	b.WriteString("======\n\n")
	fmt.Fprintf(&b, "Room %s  [%s]  %s\n", snapshot.RoomID, snapshot.Status, snapshot.TimeControl.String())
	fmt.Fprintf(&b, "You: %s (%s)\n", nickname, roleLabel(role))
	fmt.Fprintf(&b, "White: %s  %s\n", fallback(snapshot.WhiteName, "open"), formatDuration(snapshot.WhiteTimeLeft))
	fmt.Fprintf(&b, "Black: %s  %s\n", fallback(snapshot.BlackName, "open"), formatDuration(snapshot.BlackTimeLeft))
	fmt.Fprintf(&b, "Watchers: %d\n", snapshot.WatcherCount)

	if snapshot.Outcome != "" {
		fmt.Fprintf(&b, "Result: %s by %s\n", snapshot.Outcome, snapshot.Method)
	} else {
		fmt.Fprintf(&b, "Turn: %s\n", strings.ToLower(snapshot.Turn))
	}

	if snapshot.LastEvent != "" {
		fmt.Fprintf(&b, "Event: %s\n", snapshot.LastEvent)
	}

	b.WriteString("\n")
	b.WriteString(snapshot.Board)
	b.WriteString("\n")
	b.WriteString(renderMoves(snapshot.Moves))
	b.WriteString("\nCommands\n")
	b.WriteString("  help                show help\n")
	b.WriteString("  leave               leave the room\n")
	b.WriteString("  resign              resign if you are seated in an active game\n")
	b.WriteString("  board               redraw the board\n")
	if role == domain.RoleWhite || role == domain.RoleBlack {
		b.WriteString("  <SAN move>          play a move like e4, Nf3, O-O, Qxe5+\n")
	} else {
		b.WriteString("  Spectators cannot move, but they receive live updates.\n")
	}
	return b.String()
}

func HelpText(inRoom bool) string {
	if inRoom {
		return "Room commands: help, board, leave, resign, or enter a SAN move like e4 or O-O."
	}
	return "Lobby commands: list, create <tc>, join <id>, watch <id>, help, quit."
}

func Prompt(snapshot *domain.GameSnapshot, role domain.Role) string {
	if snapshot == nil {
		return "lobby> "
	}

	if snapshot.Status == domain.RoomStatusActive {
		if role == domain.RoleWhite && strings.EqualFold(snapshot.Turn, "white") {
			return "move> "
		}
		if role == domain.RoleBlack && strings.EqualFold(snapshot.Turn, "black") {
			return "move> "
		}
	}
	return fmt.Sprintf("room[%s]> ", snapshot.RoomID)
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

func renderMoves(moves []string) string {
	if len(moves) == 0 {
		return "Moves: none yet\n"
	}

	start := 0
	if len(moves) > 12 {
		start = len(moves) - 12
		if start%2 != 0 {
			start--
		}
	}

	var b strings.Builder
	b.WriteString("Recent moves\n")
	for i := start; i < len(moves); i += 2 {
		moveNo := i/2 + 1
		white := moves[i]
		black := ""
		if i+1 < len(moves) {
			black = moves[i+1]
		}
		fmt.Fprintf(&b, "  %2d. %-8s %s\n", moveNo, white, black)
	}
	return b.String()
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
