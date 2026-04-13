package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Description string

	Title    lipgloss.Style
	Section  lipgloss.Style
	Accent   lipgloss.Style
	Emphasis lipgloss.Style
	Muted    lipgloss.Style
	Dim      lipgloss.Style

	StatusInfo lipgloss.Style
	StatusOK   lipgloss.Style
	StatusWarn lipgloss.Style
	StatusErr  lipgloss.Style

	BadgeWaiting  lipgloss.Style
	BadgeActive   lipgloss.Style
	BadgeFinished lipgloss.Style

	WhitePiece lipgloss.Style
	BlackPiece lipgloss.Style

	LightSquare   lipgloss.Style
	DarkSquare    lipgloss.Style
	LastMoveLight lipgloss.Style
	LastMoveDark  lipgloss.Style
	CheckSquare   lipgloss.Style
	CoordLabel    lipgloss.Style

	ClockNormal lipgloss.Style
	ClockActive lipgloss.Style
	ClockWarn   lipgloss.Style
	ClockAlert  lipgloss.Style

	MoveDefault lipgloss.Style
	MoveLatest  lipgloss.Style
	MoveCheck   lipgloss.Style
	MoveMate    lipgloss.Style
	MoveCapture lipgloss.Style
	MoveCastle  lipgloss.Style

	PanelBorder lipgloss.Style
	BorderStyle lipgloss.Border

	EmptyLight rune
	EmptyDark  rune
}
