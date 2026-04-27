package theme

import "github.com/charmbracelet/lipgloss"

func Classic() Theme {
	c := func(s string) lipgloss.Color { return lipgloss.Color(s) }

	border := lipgloss.RoundedBorder()
	panelBorder := lipgloss.NewStyle().Foreground(c("65"))

	return Theme{
		Name:        "classic",
		Description: "The original e4 palette: green title, olive highlights, deep board.",

		Title:    lipgloss.NewStyle().Bold(true).Foreground(c("120")),
		Section:  lipgloss.NewStyle().Bold(true).Foreground(c("114")),
		Accent:   lipgloss.NewStyle().Bold(true).Foreground(c("186")),
		Emphasis: lipgloss.NewStyle().Bold(true).Foreground(c("229")),
		Muted:    lipgloss.NewStyle().Foreground(c("71")),
		Dim:      lipgloss.NewStyle().Faint(true).Foreground(c("65")),

		StatusInfo: lipgloss.NewStyle().Bold(true).Foreground(c("123")),
		StatusOK:   lipgloss.NewStyle().Bold(true).Foreground(c("120")),
		StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(c("186")),
		StatusErr:  lipgloss.NewStyle().Bold(true).Foreground(c("203")),

		BadgeWaiting:  lipgloss.NewStyle().Bold(true).Foreground(c("16")).Background(c("180")).Padding(0, 1),
		BadgeActive:   lipgloss.NewStyle().Bold(true).Foreground(c("16")).Background(c("114")).Padding(0, 1),
		BadgeFinished: lipgloss.NewStyle().Bold(true).Foreground(c("16")).Background(c("110")).Padding(0, 1),

		WhitePiece: lipgloss.NewStyle().Bold(true).Foreground(c("231")),
		BlackPiece: lipgloss.NewStyle().Bold(true).Foreground(c("117")),

		LightSquare:   lipgloss.NewStyle().Background(c("239")),
		DarkSquare:    lipgloss.NewStyle().Background(c("236")),
		LastMoveLight: lipgloss.NewStyle().Background(c("100")),
		LastMoveDark:  lipgloss.NewStyle().Background(c("58")),
		CheckSquare:   lipgloss.NewStyle().Background(c("88")),
		CoordLabel:    lipgloss.NewStyle().Faint(true).Foreground(c("65")),

		ClockNormal: lipgloss.NewStyle().Foreground(c("120")),
		ClockActive: lipgloss.NewStyle().Bold(true).Foreground(c("120")),
		ClockWarn:   lipgloss.NewStyle().Bold(true).Foreground(c("186")),
		ClockAlert:  lipgloss.NewStyle().Bold(true).Foreground(c("203")),

		MoveDefault: lipgloss.NewStyle().Foreground(c("151")),
		MoveLatest:  lipgloss.NewStyle().Bold(true).Foreground(c("229")),
		MoveCheck:   lipgloss.NewStyle().Foreground(c("203")),
		MoveMate:    lipgloss.NewStyle().Bold(true).Foreground(c("16")).Background(c("203")),
		MoveCapture: lipgloss.NewStyle().Foreground(c("186")),
		MoveCastle:  lipgloss.NewStyle().Foreground(c("123")),

		PanelBorder: panelBorder,
		BorderStyle: border,

		EmptyLight: ' ',
		EmptyDark:  ' ',
	}
}
