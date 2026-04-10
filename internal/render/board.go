package render

import (
	"fmt"
	"strings"

	"github.com/morum/e4/internal/domain"
)

func renderBoard(ctx Context, board domain.BoardState) []string {
	t := newTheme(ctx.ANSI)
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	ranks := []int{8, 7, 6, 5, 4, 3, 2, 1}
	if ctx.Orientation == OrientationBlack {
		files = []string{"h", "g", "f", "e", "d", "c", "b", "a"}
		ranks = []int{1, 2, 3, 4, 5, 6, 7, 8}
	}

	coordLine := "    " + strings.Join(files, "  ")
	lines := []string{t.dim(coordLine)}
	for _, rank := range ranks {
		var row strings.Builder
		row.WriteString(t.dim(fmt.Sprintf(" %d ", rank)))
		for _, file := range files {
			square := fmt.Sprintf("%s%d", file, rank)
			row.WriteString(renderBoardCell(t, board, square))
		}
		row.WriteString(t.dim(fmt.Sprintf(" %d", rank)))
		lines = append(lines, row.String())
	}
	lines = append(lines, t.dim(coordLine))
	return lines
}

func renderBoardCell(t theme, board domain.BoardState, square string) string {
	piece, occupied := board.Squares[square]
	content := squareFill(square)
	styles := []string{cellBackground(square)}

	if square == board.LastMoveFrom || square == board.LastMoveTo {
		styles[0] = bg256(58)
	}
	if square == board.CheckSquare {
		styles[0] = bg256(52)
	}

	if occupied {
		content = piece.Symbol
		if piece.Color == "white" {
			styles = append(styles, "1", fg256(255))
		} else {
			styles = append(styles, "1", fg256(117))
		}
	} else {
		styles = append(styles, fg256(72))
	}

	return t.paint(" "+content+" ", styles...)
}

func cellBackground(square string) string {
	file := int(square[0] - 'a')
	rank := int(square[1] - '1')
	if (file+rank)%2 == 0 {
		return bg256(233)
	}
	return bg256(236)
}

func squareFill(square string) string {
	file := int(square[0] - 'a')
	rank := int(square[1] - '1')
	if (file+rank)%2 == 0 {
		return "."
	}
	return ":"
}
