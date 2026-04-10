package service

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"chessh/internal/domain"
)

type RoomRepository interface {
	Save(room GameRoom) error
	Get(id string) (GameRoom, bool)
	Delete(id string)
	List() []GameRoom
}

type LobbyService struct {
	repo   RoomRepository
	logger *slog.Logger
}

func NewLobbyService(repo RoomRepository, logger *slog.Logger) *LobbyService {
	return &LobbyService{repo: repo, logger: logger}
}

func (s *LobbyService) CreateGame(host domain.Participant, tc domain.TimeControl) (GameRoom, domain.Role, error) {
	for range 32 {
		id, err := generateRoomID()
		if err != nil {
			return nil, domain.RoleNone, err
		}

		if _, exists := s.repo.Get(id); exists {
			continue
		}

		room := NewRoom(id, tc, s.logger)
		if err := s.repo.Save(room); err != nil {
			return nil, domain.RoleNone, err
		}

		role, err := room.JoinPlayer(host)
		if err != nil {
			s.repo.Delete(id)
			return nil, domain.RoleNone, err
		}

		s.logInfo("room created", "room_id", id, "time_control", tc.String(), "session_id", host.ID, "nickname", host.Nickname)

		return room, role, nil
	}

	return nil, domain.RoleNone, fmt.Errorf("failed to generate a unique room ID")
}

func (s *LobbyService) ListGames() []domain.RoomSummary {
	rooms := s.repo.List()
	summaries := make([]domain.RoomSummary, 0, len(rooms))
	for _, room := range rooms {
		snapshot := room.Snapshot()
		if snapshot.ParticipantCount() == 0 {
			continue
		}
		summaries = append(summaries, snapshot.Summary())
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Status != summaries[j].Status {
			return lobbyStatusRank(summaries[i].Status) < lobbyStatusRank(summaries[j].Status)
		}
		return strings.Compare(summaries[i].ID, summaries[j].ID) < 0
	})

	return summaries
}

func (s *LobbyService) JoinGame(id string, participant domain.Participant) (GameRoom, domain.Role, error) {
	room, ok := s.repo.Get(id)
	if !ok {
		return nil, domain.RoleNone, fmt.Errorf("room %s was not found", id)
	}

	role, err := room.JoinPlayer(participant)
	if err != nil {
		return nil, domain.RoleNone, err
	}

	s.logInfo("room joined", "room_id", id, "session_id", participant.ID, "nickname", participant.Nickname, "role", role)

	return room, role, nil
}

func (s *LobbyService) WatchGame(id string, participant domain.Participant) (GameRoom, error) {
	room, ok := s.repo.Get(id)
	if !ok {
		return nil, fmt.Errorf("room %s was not found", id)
	}

	if err := room.AddWatcher(participant); err != nil {
		return nil, err
	}

	s.logInfo("room watched", "room_id", id, "session_id", participant.ID, "nickname", participant.Nickname)

	return room, nil
}

func (s *LobbyService) LeaveRoom(id, participantID string) error {
	room, ok := s.repo.Get(id)
	if !ok {
		return nil
	}

	if room.Leave(participantID) {
		s.repo.Delete(id)
		s.logInfo("room removed", "room_id", id)
	}
	s.logInfo("room left", "room_id", id, "session_id", participantID)

	return nil
}

func (s *LobbyService) logInfo(msg string, attrs ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Info(msg, attrs...)
}

func lobbyStatusRank(status domain.RoomStatus) int {
	switch status {
	case domain.RoomStatusActive:
		return 0
	case domain.RoomStatusWaiting:
		return 1
	default:
		return 2
	}
}

func generateRoomID() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}
