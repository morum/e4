package theme

import "github.com/charmbracelet/lipgloss"

func Mono() Theme {
	plain := lipgloss.NewStyle()
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Faint(true)
	reverse := lipgloss.NewStyle().Reverse(true)

	return Theme{
		Name:        "mono",
		Description: "Monochrome theme using only bold, dim, and reverse-video. Always readable.",

		Title:    bold,
		Section:  bold,
		Accent:   bold,
		Emphasis: bold,
		Muted:    plain,
		Dim:      dim,

		StatusInfo: bold,
		StatusOK:   bold,
		StatusWarn: bold,
		StatusErr:  bold.Reverse(true),

		BadgeWaiting:  reverse.Padding(0, 1),
		BadgeActive:   reverse.Bold(true).Padding(0, 1),
		BadgeFinished: dim.Padding(0, 1),

		WhitePiece: bold,
		BlackPiece: bold.Faint(true),

		LightSquare:   plain,
		DarkSquare:    reverse,
		LastMoveLight: bold.Underline(true),
		LastMoveDark:  reverse.Bold(true).Underline(true),
		CheckSquare:   reverse.Bold(true),
		CoordLabel:    dim,

		ClockNormal: plain,
		ClockActive: bold,
		ClockWarn:   bold,
		ClockAlert:  bold.Reverse(true),

		MoveDefault: plain,
		MoveLatest:  bold,
		MoveCheck:   bold,
		MoveMate:    bold.Reverse(true),
		MoveCapture: bold,
		MoveCastle:  bold,

		PanelBorder: dim,
		BorderStyle: lipgloss.NormalBorder(),

		EmptyLight: ' ',
		EmptyDark:  ' ',
	}
}
