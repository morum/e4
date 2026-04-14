package widget

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// WrapPieces joins pieces with sep, breaking to a new line whenever the next
// piece would push the running width past maxW. A very small maxW still emits
// one piece per line — we never truncate the piece itself.
func WrapPieces(pieces []string, sep string, maxW int) string {
	if maxW <= 0 {
		return strings.Join(pieces, sep)
	}
	sepW := lipgloss.Width(sep)
	var lines []string
	var current strings.Builder
	currentW := 0
	for _, p := range pieces {
		pw := lipgloss.Width(p)
		switch {
		case currentW == 0:
			current.WriteString(p)
			currentW = pw
		case currentW+sepW+pw > maxW:
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(p)
			currentW = pw
		default:
			current.WriteString(sep)
			current.WriteString(p)
			currentW += sepW + pw
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}

// TruncateLine clips a plain (unstyled) line to maxW visible columns. Lines
// that already contain ANSI escapes are returned as-is to avoid splitting
// a color sequence — for those, callers should prefer WrapPieces.
func TruncateLine(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	if strings.Contains(s, "\x1b") {
		return s
	}
	if maxW == 1 {
		return "…"
	}
	return s[:maxW-1] + "…"
}
