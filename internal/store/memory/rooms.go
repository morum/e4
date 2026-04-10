package memory

import (
	"fmt"
	"sync"

	"chessh/internal/service"
)

type RoomRepository struct {
	mu    sync.RWMutex
	rooms map[string]service.GameRoom
}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{rooms: make(map[string]service.GameRoom)}
}

func (r *RoomRepository) Save(room service.GameRoom) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rooms[room.ID()]; exists {
		return fmt.Errorf("room %s already exists", room.ID())
	}
	r.rooms[room.ID()] = room
	return nil
}

func (r *RoomRepository) Get(id string) (service.GameRoom, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, ok := r.rooms[id]
	return room, ok
}

func (r *RoomRepository) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rooms, id)
}

func (r *RoomRepository) List() []service.GameRoom {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rooms := make([]service.GameRoom, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
