package widget

import (
	"strings"

	"github.com/morum/e4/internal/tui/theme"
)

var bannerLines = []string{
	`                 _  _   `,
	`     ___    ___ | || |  `,
	`    / -_)  |___|| || |  `,
	`    \___|       |_||_|  `,
	`                        `,
}

func Banner(t theme.Theme) string {
	out := make([]string, 0, len(bannerLines)+2)
	for _, line := range bannerLines {
		out = append(out, t.Title.Render(line))
	}
	out = append(out, t.Muted.Render("        terminal chess over ssh"))
	return strings.Join(out, "\n")
}
