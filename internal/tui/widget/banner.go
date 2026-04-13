package widget

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/tui/theme"
)

var bannerLines = []string{
	`     ___ `,
	` ___| | |`,
	`| -_|_  |`,
	`|___| |_|`,
}

func Banner(t theme.Theme) string {
	parts := make([]string, 0, len(bannerLines)+2)
	for _, line := range bannerLines {
		parts = append(parts, t.Title.Render(line))
	}
	parts = append(parts, "", t.Muted.Render("terminal chess over ssh"))
	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}
