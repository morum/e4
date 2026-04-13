package widget

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/tui/theme"
)

func MoveList(t theme.Theme, moves []string, width, height int) string {
	if width < 12 {
		width = 12
	}
	if height < 3 {
		height = 3
	}

	header := t.Section.Render("Recent Moves")
	rule := t.PanelBorder.Render(strings.Repeat("─", width))

	body := buildMoveBody(t, moves, height-2)
	return strings.Join([]string{header, rule, body}, "\n")
}

func buildMoveBody(t theme.Theme, moves []string, height int) string {
	if len(moves) == 0 {
		return t.Dim.Render("  none yet")
	}

	pairs := (len(moves) + 1) / 2
	maxRows := max(height, 1)

	startPair := 0
	if pairs > maxRows {
		startPair = pairs - maxRows
	}

	var lines []string
	for p := startPair; p < pairs; p++ {
		i := p * 2
		latest := false
		whiteMove := styleMove(t, moves[i], i == len(moves)-1)
		blackMove := ""
		if i+1 < len(moves) {
			latest = i+1 == len(moves)-1
			blackMove = styleMove(t, moves[i+1], latest)
		}
		num := t.Dim.Render(fmt.Sprintf("%3d.", p+1))
		line := fmt.Sprintf("%s %s %s", num, padMove(whiteMove, 8), padMove(blackMove, 8))
		lines = append(lines, line)
	}

	for len(lines) < maxRows {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func styleMove(t theme.Theme, move string, latest bool) string {
	style := t.MoveDefault
	switch {
	case strings.Contains(move, "#"):
		style = t.MoveMate
	case strings.Contains(move, "+"):
		style = t.MoveCheck
	case strings.HasPrefix(move, "O-O"):
		style = t.MoveCastle
	case strings.Contains(move, "x"):
		style = t.MoveCapture
	case latest:
		style = t.MoveLatest
	}
	return style.Render(move)
}

func padMove(m string, width int) string {
	w := lipgloss.Width(m)
	if w >= width {
		return m
	}
	return m + strings.Repeat(" ", width-w)
}
