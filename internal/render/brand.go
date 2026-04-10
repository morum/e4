package render

import "strings"

func JoinBanner(ansi bool) string {
	t := newTheme(ansi)
	lines := []string{
		t.title("███████╗██╗  ██╗"),
		t.title("██╔════╝██║  ██║"),
		t.title("█████╗  ███████║"),
		t.title("██╔══╝  ╚════██║"),
		t.title("███████╗     ██║"),
		t.title("╚══════╝     ╚═╝"),
		"",
		t.accent("terminal chess over ssh"),
		t.muted("Enter a nickname to get started."),
	}

	return strings.Join(lines, "\n")
}
