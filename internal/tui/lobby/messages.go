package lobby

import (
	"github.com/morum/e4/internal/domain"
	"github.com/morum/e4/internal/service"
)

type roomsLoadedMsg struct {
	Rooms []domain.RoomSummary
}

type roomsErrorMsg struct {
	Err error
}

type refreshTickMsg struct{}

type joinResultMsg struct {
	Room service.GameRoom
	Role domain.Role
	Sub  service.RoomSubscription
	Err  error
}

// EnterRoomMsg signals the parent app that the user is ready to attach to a room.
type EnterRoomMsg struct {
	Room service.GameRoom
	Role domain.Role
	Sub  service.RoomSubscription
}

// CycleThemeMsg signals the parent app to advance to the next theme.
type CycleThemeMsg struct{}

// SetThemeMsg requests the parent app set a specific theme.
type SetThemeMsg struct{ Name string }
