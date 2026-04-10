package service

import (
	"strings"

	"chessh/internal/domain"

	"github.com/notnil/chess"
)

func buildBoardState(position *chess.Position, lastMoveFrom, lastMoveTo string) domain.BoardState {
	board := position.Board()
	squareMap := board.SquareMap()
	squares := make(map[string]domain.BoardPiece, len(squareMap))
	for square, piece := range squareMap {
		squares[square.String()] = domain.BoardPiece{
			Color:  strings.ToLower(piece.Color().Name()),
			Symbol: pieceSymbol(piece),
		}
	}

	return domain.BoardState{
		Squares:      squares,
		LastMoveFrom: lastMoveFrom,
		LastMoveTo:   lastMoveTo,
		CheckSquare:  checkedKingSquare(position),
	}
}

func checkedKingSquare(position *chess.Position) string {
	turn := position.Turn()
	kingSquare, ok := findKingSquare(position.Board().SquareMap(), turn)
	if !ok {
		return ""
	}

	if isSquareAttacked(position.Board().SquareMap(), kingSquare, turn.Other()) {
		return kingSquare.String()
	}

	return ""
}

func findKingSquare(squareMap map[chess.Square]chess.Piece, color chess.Color) (chess.Square, bool) {
	for square, piece := range squareMap {
		if piece.Color() == color && piece.Type() == chess.King {
			return square, true
		}
	}

	return chess.NoSquare, false
}

func isSquareAttacked(squareMap map[chess.Square]chess.Piece, target chess.Square, attackerColor chess.Color) bool {
	for square, piece := range squareMap {
		if piece.Color() != attackerColor {
			continue
		}

		if pieceAttacksSquare(squareMap, square, target, piece) {
			return true
		}
	}

	return false
}

func pieceAttacksSquare(squareMap map[chess.Square]chess.Piece, from, target chess.Square, piece chess.Piece) bool {
	fileDelta := int(target.File()) - int(from.File())
	rankDelta := int(target.Rank()) - int(from.Rank())
	absFile := abs(fileDelta)
	absRank := abs(rankDelta)

	switch piece.Type() {
	case chess.Pawn:
		return pawnAttacksSquare(piece.Color(), fileDelta, rankDelta)
	case chess.Knight:
		return (absFile == 1 && absRank == 2) || (absFile == 2 && absRank == 1)
	case chess.Bishop:
		return absFile == absRank && pathClear(squareMap, from, target, sign(fileDelta), sign(rankDelta))
	case chess.Rook:
		if fileDelta != 0 && rankDelta != 0 {
			return false
		}
		return pathClear(squareMap, from, target, sign(fileDelta), sign(rankDelta))
	case chess.Queen:
		if absFile == absRank {
			return pathClear(squareMap, from, target, sign(fileDelta), sign(rankDelta))
		}
		if fileDelta == 0 || rankDelta == 0 {
			return pathClear(squareMap, from, target, sign(fileDelta), sign(rankDelta))
		}
		return false
	case chess.King:
		return absFile <= 1 && absRank <= 1
	default:
		return false
	}
}

func pawnAttacksSquare(color chess.Color, fileDelta, rankDelta int) bool {
	if abs(fileDelta) != 1 {
		return false
	}

	if color == chess.White {
		return rankDelta == 1
	}

	if color == chess.Black {
		return rankDelta == -1
	}

	return false
}

func pathClear(squareMap map[chess.Square]chess.Piece, from, target chess.Square, fileStep, rankStep int) bool {
	if fileStep == 0 && rankStep == 0 {
		return false
	}

	file := int(from.File()) + fileStep
	rank := int(from.Rank()) + rankStep
	targetFile := int(target.File())
	targetRank := int(target.Rank())

	for file != targetFile || rank != targetRank {
		square := chess.NewSquare(chess.File(file), chess.Rank(rank))
		if _, blocked := squareMap[square]; blocked {
			return false
		}
		file += fileStep
		rank += rankStep
	}

	return true
}

func pieceSymbol(piece chess.Piece) string {
	symbol := strings.ToUpper(piece.Type().String())
	if piece.Color() == chess.Black {
		return strings.ToLower(symbol)
	}
	return symbol
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
