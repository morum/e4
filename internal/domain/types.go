package domain

import "time"

type Role string

const (
	RoleNone    Role = "none"
	RoleWhite   Role = "white"
	RoleBlack   Role = "black"
	RoleWatcher Role = "watcher"
)

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusActive   RoomStatus = "active"
	RoomStatusFinished RoomStatus = "finished"
)

type Participant struct {
	ID       string
	Nickname string
}

type RoomSummary struct {
	ID           string
	Status       RoomStatus
	TimeControl  TimeControl
	WhiteName    string
	BlackName    string
	WatcherCount int
	HasOpenSeat  bool
	Turn         string
	Outcome      string
	Method       string
}

type BoardPiece struct {
	Color  string
	Symbol string
}

type BoardState struct {
	Squares      map[string]BoardPiece
	LastMoveFrom string
	LastMoveTo   string
	CheckSquare  string
}

type GameSnapshot struct {
	RoomID        string
	Status        RoomStatus
	TimeControl   TimeControl
	WhiteID       string
	WhiteName     string
	BlackID       string
	BlackName     string
	WatcherCount  int
	Turn          string
	Board         BoardState
	Moves         []string
	WhiteTimeLeft time.Duration
	BlackTimeLeft time.Duration
	Outcome       string
	Method        string
	LastEvent     string
}

func (s GameSnapshot) Summary() RoomSummary {
	return RoomSummary{
		ID:           s.RoomID,
		Status:       s.Status,
		TimeControl:  s.TimeControl,
		WhiteName:    s.WhiteName,
		BlackName:    s.BlackName,
		WatcherCount: s.WatcherCount,
		HasOpenSeat:  s.IsSeatOpen(),
		Turn:         s.Turn,
		Outcome:      s.Outcome,
		Method:       s.Method,
	}
}

func (s GameSnapshot) IsSeatOpen() bool {
	return s.Status == RoomStatusWaiting && (s.WhiteID == "" || s.BlackID == "")
}

func (s GameSnapshot) ParticipantCount() int {
	count := s.WatcherCount
	if s.WhiteID != "" {
		count++
	}
	if s.BlackID != "" {
		count++
	}
	return count
}
