package room

import (
	"time"

	"github.com/morum/e4/internal/domain"
)

type SnapshotMsg domain.GameSnapshot

type subscriptionClosedMsg struct{}

type tickMsg time.Time

type moveSubmittedMsg struct {
	Move string
	Err  error
}

type resignedMsg struct {
	Err error
}

// LeaveRequestMsg signals the parent app that the user wants to return to the lobby.
type LeaveRequestMsg struct {
	Reason string
}

// CycleThemeMsg signals the parent app to advance to the next theme.
type CycleThemeMsg struct{}
