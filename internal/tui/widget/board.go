package widget

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/tui/theme"
)

type BoardSize int

const (
	BoardSmall BoardSize = iota
	BoardMedium
	BoardLarge
)

type cellGeom struct{ w, h int }

var boardGeoms = map[BoardSize]cellGeom{
	BoardSmall:  {3, 1},
	BoardMedium: {5, 3},
	BoardLarge:  {7, 3},
}

const rankGutter = 3

func (s BoardSize) Cell() (w, h int) { g := boardGeoms[s]; return g.w, g.h }

func (s BoardSize) Footprint() (w, h int) {
	g := boardGeoms[s]
	return g.w*8 + 2*rankGutter, g.h*8 + 2
}

func PickBoardSize(width, height int) BoardSize {
	for _, sz := range []BoardSize{BoardLarge, BoardMedium, BoardSmall} {
		fw, fh := sz.Footprint()
		if width >= fw && height >= fh {
			return sz
		}
	}
	return BoardSmall
}

type BoardOptions struct {
	Size        BoardSize
	Orientation domain.Role
}

func Board(t theme.Theme, board domain.BoardState, opts BoardOptions) string {
	cellW, cellH := opts.Size.Cell()
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	ranks := []int{8, 7, 6, 5, 4, 3, 2, 1}
	if opts.Orientation == domain.RoleBlack {
		files = []string{"h", "g", "f", "e", "d", "c", "b", "a"}
		ranks = []int{1, 2, 3, 4, 5, 6, 7, 8}
	}

	fileRow := buildFileLabelRow(t, files, cellW)
	lines := []string{fileRow}
	for _, rank := range ranks {
		lines = append(lines, renderRank(t, board, files, rank, cellW, cellH)...)
	}
	lines = append(lines, fileRow)
	return strings.Join(lines, "\n")
}

func buildFileLabelRow(t theme.Theme, files []string, cellW int) string {
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", rankGutter))
	for _, f := range files {
		b.WriteString(t.CoordLabel.Render(centerText(f, cellW)))
	}
	b.WriteString(strings.Repeat(" ", rankGutter))
	return b.String()
}

func renderRank(t theme.Theme, board domain.BoardState, files []string, rank, cellW, cellH int) []string {
	cells := make([]string, len(files))
	for i, f := range files {
		cells[i] = renderSquare(t, board, f, rank, cellW, cellH)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	rows := strings.Split(row, "\n")

	mid := cellH / 2
	rankLabel := t.CoordLabel.Render(rankLabelText(rank))
	gutter := strings.Repeat(" ", rankGutter)

	out := make([]string, len(rows))
	for i, r := range rows {
		prefix, suffix := gutter, gutter
		if i == mid {
			prefix = rankLabel
			suffix = rankLabel
		}
		out[i] = prefix + r + suffix
	}
	return out
}

func rankLabelText(rank int) string {
	if rank < 1 || rank > 8 {
		return "   "
	}
	return " " + string(rune('0'+rank)) + " "
}

func renderSquare(t theme.Theme, board domain.BoardState, file string, rank, cellW, cellH int) string {
	square := file + string(rune('0'+rank))
	light := isLightSquare(file, rank)

	bg := t.DarkSquare
	if light {
		bg = t.LightSquare
	}
	if square == board.LastMoveFrom || square == board.LastMoveTo {
		bg = t.LastMoveDark
		if light {
			bg = t.LastMoveLight
		}
	}
	if square == board.CheckSquare {
		bg = t.CheckSquare
	}

	piece, occupied := board.Squares[square]
	mid := cellH / 2

	rows := make([]string, cellH)
	for i := 0; i < cellH; i++ {
		blank := strings.Repeat(" ", cellW)
		if i == mid && occupied {
			content := centerText(piece.Symbol, cellW)
			pieceStyle := t.WhitePiece
			if piece.Color == "black" {
				pieceStyle = t.BlackPiece
			}
			pieceStyle = pieceStyle.Background(bg.GetBackground())
			rows[i] = pieceStyle.Render(content)
			continue
		}
		rows[i] = bg.Render(blank)
	}
	return strings.Join(rows, "\n")
}

func isLightSquare(file string, rank int) bool {
	if len(file) == 0 {
		return false
	}
	f := int(file[0] - 'a')
	return (f+rank)%2 == 1
}

func centerText(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	pad := width - w
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
