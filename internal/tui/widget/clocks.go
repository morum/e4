package widget

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/tui/theme"
)

type ClockSide struct {
	Label    string
	Name     string
	Time     time.Duration
	Active   bool
	InCheck  bool
	IsBottom bool
}

func ClockPair(t theme.Theme, top, bottom ClockSide, width int) string {
	return strings.Join([]string{
		clockLine(t, top, width),
		clockLine(t, bottom, width),
	}, "\n")
}

func clockLine(t theme.Theme, side ClockSide, width int) string {
	marker := "  "
	if side.Active {
		marker = t.Accent.Render("▶ ")
	}

	name := side.Name
	if strings.TrimSpace(name) == "" {
		name = "(open)"
	}
	nameStyled := t.Emphasis.Render(side.Label) + ": " + t.Accent.Render(name)
	if side.InCheck {
		nameStyled += " " + t.StatusErr.Render("CHECK")
	}

	clockStyle := t.ClockNormal
	switch {
	case side.Time <= 10*time.Second:
		clockStyle = t.ClockAlert
	case side.Time <= 30*time.Second:
		clockStyle = t.ClockWarn
	case side.Active:
		clockStyle = t.ClockActive
	}
	clockText := clockStyle.Render(formatDuration(side.Time))

	left := marker + nameStyled
	right := clockText

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := width - leftW - rightW
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
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
