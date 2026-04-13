package theme

import "github.com/charmbracelet/lipgloss"

func NightOwl() Theme {
	c := func(s string) lipgloss.Color { return lipgloss.Color(s) }

	bg := c("#011627")
	bgAlt := c("#1d3b53")
	fg := c("#d6deeb")
	muted := c("#637777")
	cyan := c("#7fdbca")
	yellow := c("#ecc48d")
	pink := c("#c792ea")
	red := c("#ef5350")
	green := c("#22da6e")
	orange := c("#f78c6c")

	light := c("#3b536b")
	dark := c("#1d3b53")
	highlight := c("#7e57c2")
	highlightDark := c("#5a3a91")
	check := c("#83202e")

	return Theme{
		Name:        "nightowl",
		Description: "High-contrast dark theme inspired by Sarah Drasner's Night Owl.",

		Title:    lipgloss.NewStyle().Bold(true).Foreground(cyan),
		Section:  lipgloss.NewStyle().Bold(true).Foreground(yellow),
		Accent:   lipgloss.NewStyle().Bold(true).Foreground(pink),
		Emphasis: lipgloss.NewStyle().Bold(true).Foreground(orange),
		Muted:    lipgloss.NewStyle().Foreground(muted),
		Dim:      lipgloss.NewStyle().Faint(true).Foreground(muted),

		StatusInfo: lipgloss.NewStyle().Bold(true).Foreground(cyan),
		StatusOK:   lipgloss.NewStyle().Bold(true).Foreground(green),
		StatusWarn: lipgloss.NewStyle().Bold(true).Foreground(yellow),
		StatusErr:  lipgloss.NewStyle().Bold(true).Foreground(red),

		BadgeWaiting:  lipgloss.NewStyle().Bold(true).Foreground(bg).Background(orange).Padding(0, 1),
		BadgeActive:   lipgloss.NewStyle().Bold(true).Foreground(bg).Background(green).Padding(0, 1),
		BadgeFinished: lipgloss.NewStyle().Bold(true).Foreground(bg).Background(muted).Padding(0, 1),

		WhitePiece: lipgloss.NewStyle().Bold(true).Foreground(fg),
		BlackPiece: lipgloss.NewStyle().Bold(true).Foreground(cyan),

		LightSquare:   lipgloss.NewStyle().Background(light),
		DarkSquare:    lipgloss.NewStyle().Background(dark),
		LastMoveLight: lipgloss.NewStyle().Background(highlight),
		LastMoveDark:  lipgloss.NewStyle().Background(highlightDark),
		CheckSquare:   lipgloss.NewStyle().Background(check),
		CoordLabel:    lipgloss.NewStyle().Faint(true).Foreground(muted),

		ClockNormal: lipgloss.NewStyle().Foreground(fg),
		ClockActive: lipgloss.NewStyle().Bold(true).Foreground(green),
		ClockWarn:   lipgloss.NewStyle().Bold(true).Foreground(yellow),
		ClockAlert:  lipgloss.NewStyle().Bold(true).Foreground(red),

		MoveDefault: lipgloss.NewStyle().Foreground(fg),
		MoveLatest:  lipgloss.NewStyle().Bold(true).Foreground(yellow),
		MoveCheck:   lipgloss.NewStyle().Foreground(red),
		MoveMate:    lipgloss.NewStyle().Bold(true).Foreground(bg).Background(red),
		MoveCapture: lipgloss.NewStyle().Foreground(orange),
		MoveCastle:  lipgloss.NewStyle().Foreground(pink),

		PanelBorder: lipgloss.NewStyle().Foreground(bgAlt),
		BorderStyle: lipgloss.RoundedBorder(),

		EmptyLight: ' ',
		EmptyDark:  ' ',
	}
}
