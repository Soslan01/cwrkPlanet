package ws

import (
	"sync"
)

type Conn interface {
	Send(msg Message) error
	Close() error
	UserID() string
	RoomID() string
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[Conn]struct{} // roomID -> set of connections
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[Conn]struct{})}
}

func (h *Hub) Add(c Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	rs, ok := h.rooms[c.RoomID()]
	if !ok {
		rs = make(map[Conn]struct{})
		h.rooms[c.RoomID()] = rs
	}
	rs[c] = struct{}{}
}

func (h *Hub) Remove(c Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if rs, ok := h.rooms[c.RoomID()]; ok {
		delete(rs, c)
		if len(rs) == 0 {
			delete(h.rooms, c.RoomID())
		}
	}
}

func (h *Hub) Broadcast(roomID string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if rs, ok := h.rooms[roomID]; ok {
		for c := range rs {
			_ = c.Send(msg) // best-effort
		}
	}
}
