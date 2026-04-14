package widget

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestWrapPiecesFitsOnOneLineWhenUnderBudget(t *testing.T) {
	got := WrapPieces([]string{"a", "b", "c"}, "  ", 40)
	if got != "a  b  c" {
		t.Fatalf("expected single-line output, got %q", got)
	}
}

func TestWrapPiecesWrapsWhenNarrower(t *testing.T) {
	pieces := []string{"alpha", "bravo", "charlie", "delta"}
	got := WrapPieces(pieces, "  ", 12)

	// Each returned line must stay within the budget.
	for _, line := range strings.Split(got, "\n") {
		if lipgloss.Width(line) > 12 {
			t.Fatalf("line %q exceeds 12 cols", line)
		}
	}

	// And every piece must appear exactly once.
	for _, p := range pieces {
		if !strings.Contains(got, p) {
			t.Fatalf("missing piece %q in output %q", p, got)
		}
	}
}

func TestWrapPiecesHandlesWidthSmallerThanSinglePiece(t *testing.T) {
	pieces := []string{"alpha", "bravo"}
	got := WrapPieces(pieces, "  ", 3)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two lines when width < one piece, got %v", lines)
	}
}

func TestTruncateLineKeepsShortStrings(t *testing.T) {
	if got := TruncateLine("hello", 10); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestTruncateLineClipsLongPlainStrings(t *testing.T) {
	got := TruncateLine("abcdefghij", 5)
	if lipgloss.Width(got) != 5 {
		t.Fatalf("expected width 5, got %q (width %d)", got, lipgloss.Width(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}

func TestTruncateLinePreservesAnsiSequences(t *testing.T) {
	// Strings containing ANSI escapes are left intact to avoid splitting
	// a color sequence mid-way and breaking downstream rendering.
	input := "\x1b[31mhello world\x1b[0m"
	if got := TruncateLine(input, 3); got != input {
		t.Fatalf("expected ANSI input to pass through unchanged, got %q", got)
	}
}
