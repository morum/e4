package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/morum/e4/internal/domain"
)

type theme struct {
	ansi bool
}

func newTheme(enabled bool) theme {
	return theme{ansi: enabled}
}

func (t theme) title(text string) string {
	return t.paint(text, "1", fg256(120))
}

func (t theme) section(text string) string {
	return t.paint(text, "1", fg256(114))
}

func (t theme) accent(text string) string {
	return t.paint(text, "1", fg256(186))
}

func (t theme) emphasis(text string) string {
	return t.paint(text, "1", fg256(229))
}

func (t theme) muted(text string) string {
	return t.paint(text, fg256(71))
}

func (t theme) dim(text string) string {
	return t.paint(text, "2", fg256(65))
}

func (t theme) role(role domain.Role, text string) string {
	switch role {
	case domain.RoleWhite:
		return t.paint(text, "1", fg256(255))
	case domain.RoleBlack:
		return t.paint(text, "1", fg256(117))
	case domain.RoleWatcher:
		return t.paint(text, "1", fg256(123))
	default:
		return text
	}
}

func (t theme) badge(status domain.RoomStatus) string {
	label := strings.ToUpper(string(status))
	switch status {
	case domain.RoomStatusWaiting:
		return t.paint(" "+label+" ", "1", fg256(16), bg256(180))
	case domain.RoomStatusActive:
		return t.paint(" "+label+" ", "1", fg256(16), bg256(114))
	case domain.RoomStatusFinished:
		return t.paint(" "+label+" ", "1", fg256(16), bg256(110))
	default:
		return label
	}
}

func (t theme) status(kind StatusKind, text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}

	prefix := "INFO"
	style := []string{"1", fg256(123)}
	switch kind {
	case StatusSuccess:
		prefix = "OK"
		style = []string{"1", fg256(120)}
	case StatusWarning:
		prefix = "WARN"
		style = []string{"1", fg256(186)}
	case StatusError:
		prefix = "ERR"
		style = []string{"1", fg256(203)}
	}

	return t.paint("["+prefix+"] ", style...) + text
}

func (t theme) clock(text string, active bool, remaining time.Duration) string {
	styles := []string{fg256(120)}
	if active {
		styles = append(styles, "1")
	}

	if remaining <= 10*time.Second {
		styles = []string{"1", fg256(203)}
	} else if remaining <= 30*time.Second {
		styles = []string{"1", fg256(186)}
	}

	return t.paint(text, styles...)
}

func (t theme) playerName(color string, text string) string {
	if color == "white" {
		return t.paint(text, "1", fg256(255))
	}
	return t.paint(text, "1", fg256(117))
}

func (t theme) move(text string, latest bool) string {
	styles := []string{}
	if latest {
		styles = append(styles, "1")
	}

	switch {
	case strings.Contains(text, "#"):
		styles = append(styles, fg256(16), bg256(203))
	case strings.Contains(text, "+"):
		styles = append(styles, fg256(203))
	case strings.Contains(text, "x"):
		styles = append(styles, fg256(186))
	case strings.HasPrefix(text, "O-O"):
		styles = append(styles, fg256(123))
	case latest:
		styles = append(styles, fg256(229))
	default:
		styles = append(styles, fg256(151))
	}

	return t.paint(text, styles...)
}

func (t theme) activeMarker(active bool) string {
	if !active {
		return t.dim("  ")
	}
	return t.accent(">>")
}

func (t theme) attention(text string) string {
	return t.paint(text, "1", fg256(203))
}

func (t theme) paint(text string, codes ...string) string {
	if !t.ansi || len(codes) == 0 {
		return text
	}
	return "\x1b[" + strings.Join(codes, ";") + "m" + text + "\x1b[0m"
}

func fg256(color int) string {
	return fmt.Sprintf("38;5;%d", color)
}

func bg256(color int) string {
	return fmt.Sprintf("48;5;%d", color)
}
