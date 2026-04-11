package render

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleWidth returns the display width of s after stripping ANSI escape sequences.
// Uses rune count rather than byte length so multi-byte characters like chess
// piece symbols (♜, ♞, etc.) are counted as single columns.
func visibleWidth(s string) int {
	return utf8.RuneCountInString(ansiEscape.ReplaceAllString(s, ""))
}

// CenterBlock horizontally centers a block of text within termWidth columns.
// All lines receive the same padding, determined by the widest visible line.
func CenterBlock(text string, termWidth int) string {
	lines := strings.Split(text, "\n")

	maxVisible := 0
	for _, line := range lines {
		if w := visibleWidth(line); w > maxVisible {
			maxVisible = w
		}
	}

	pad := 0
	if termWidth > maxVisible {
		pad = (termWidth - maxVisible) / 2
	}
	if pad <= 0 {
		return text
	}

	prefix := strings.Repeat(" ", pad)
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// centerLines pads each line so the group is visually centered within blockWidth.
// All lines get the same padding based on the widest line in the group.
func centerLines(lines []string, blockWidth int) []string {
	maxVisible := 0
	for _, line := range lines {
		if w := visibleWidth(line); w > maxVisible {
			maxVisible = w
		}
	}

	pad := 0
	if blockWidth > maxVisible {
		pad = (blockWidth - maxVisible) / 2
	}
	if pad <= 0 {
		return lines
	}

	prefix := strings.Repeat(" ", pad)
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = prefix + line
	}
	return out
}
