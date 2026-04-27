package lobby

import (
	"strings"
	"testing"

	"github.com/morum/e4/internal/domain"
)

func TestSummaryCountsTreatFilledWaitingRoomsAsActive(t *testing.T) {
	m := Model{
		rooms: []domain.RoomSummary{
			{Status: domain.RoomStatusWaiting, HasOpenSeat: true},
			{Status: domain.RoomStatusWaiting, HasOpenSeat: false},
			{Status: domain.RoomStatusFinished},
		},
	}

	got := m.summaryCounts()
	for _, want := range []string{"1 open", "1 active", "1 finished"} {
		if !strings.Contains(got, want) {
			t.Fatalf("summaryCounts() = %q, want it to contain %q", got, want)
		}
	}
}
