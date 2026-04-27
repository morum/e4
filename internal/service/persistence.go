package service

import (
	"context"
	"time"

	"github.com/morum/e4/internal/clock"
	"github.com/morum/e4/internal/domain"
)

type PlayerStore interface {
	FindOrCreateBySSHKey(ctx context.Context, key SSHKeyIdentity, nickname string) (domain.Participant, error)
}

type SSHKeyIdentity struct {
	Fingerprint   string
	AuthorizedKey string
	KeyType       string
}

type GamePersistence interface {
	CreateRoom(ctx context.Context, snapshot domain.GameSnapshot, clock clock.Snapshot) error
	UpdateRoom(ctx context.Context, snapshot domain.GameSnapshot, clock clock.Snapshot) error
	AppendMove(ctx context.Context, roomID string, move PersistedMove) error
	AppendEvent(ctx context.Context, roomID string, event PersistedEvent) error
	LoadOpenRooms(ctx context.Context) ([]PersistedRoom, error)
}

type PersistedMove struct {
	Ply            int
	PlayerID       string
	SAN            string
	From           string
	To             string
	FENAfter       string
	WhiteRemaining time.Duration
	BlackRemaining time.Duration
	PlayedAt       time.Time
}

type PersistedEvent struct {
	Type      string
	PlayerID  string
	Message   string
	CreatedAt time.Time
}

type PersistedRoom struct {
	ID          string
	Status      domain.RoomStatus
	TimeControl domain.TimeControl
	White       *domain.Participant
	Black       *domain.Participant
	Moves       []string
	Clock       clock.Snapshot
	Outcome     string
	Method      string
	LastEvent   string
	Paused      bool
}

type noopPersistence struct{}

func (noopPersistence) CreateRoom(context.Context, domain.GameSnapshot, clock.Snapshot) error {
	return nil
}
func (noopPersistence) UpdateRoom(context.Context, domain.GameSnapshot, clock.Snapshot) error {
	return nil
}
func (noopPersistence) AppendMove(context.Context, string, PersistedMove) error   { return nil }
func (noopPersistence) AppendEvent(context.Context, string, PersistedEvent) error { return nil }
func (noopPersistence) LoadOpenRooms(context.Context) ([]PersistedRoom, error)    { return nil, nil }
