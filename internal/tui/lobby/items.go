package lobby

import (
	"fmt"
	"strings"

	"github.com/morum/e4/internal/domain"
)

type roomItem struct {
	summary domain.RoomSummary
}

func (r roomItem) FilterValue() string {
	return r.summary.ID + " " + r.summary.WhiteName + " " + r.summary.BlackName
}

func (r roomItem) Title() string {
	return r.summary.ID
}

func (r roomItem) Description() string {
	return roomStateLabel(r.summary)
}

func roomStateLabel(s domain.RoomSummary) string {
	switch {
	case s.Status == domain.RoomStatusFinished:
		return fmt.Sprintf("finished — %s by %s", s.Outcome, s.Method)
	case s.Status == domain.RoomStatusActive:
		return fmt.Sprintf("active — %s to move (%s)", s.Turn, playersLabel(s))
	case s.WhiteName == "":
		return "open — white seat available"
	case s.BlackName == "":
		return "open — black seat available"
	default:
		return "waiting"
	}
}

func playersLabel(s domain.RoomSummary) string {
	white := s.WhiteName
	if strings.TrimSpace(white) == "" {
		white = "open"
	}
	black := s.BlackName
	if strings.TrimSpace(black) == "" {
		black = "open"
	}
	return fmt.Sprintf("%s vs %s", white, black)
}
