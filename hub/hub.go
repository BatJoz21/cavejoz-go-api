package hub

import (
	"sync"
	"time"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Hub struct {
	Mu          sync.RWMutex
	Connections map[int64]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		Connections: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Add(uID int64, c *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if h.Connections[uID] == nil {
		h.Connections[uID] = make(map[*Client]bool)
	}

	h.Connections[uID][c] = true
}

func (h *Hub) Send(uID int64, payload any) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	for c := range h.Connections[uID] {
		select {
		case c.Send <- payload:
			// queued
		default:
			// buffer full
		}
	}
}

func (h *Hub) Remove(uID int64, c *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	conns, ok := h.Connections[uID]
	if !ok {
		return
	}
	if _, exists := conns[c]; !exists {
		return
	}

	delete(conns, c)
	close(c.Send)

	if len(conns) == 0 {
		delete(h.Connections, uID)
	}
}

func (h *Hub) GetAll(uID int64) []*Client {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	conns := make([]*Client, 0, len(h.Connections[uID]))
	for conn := range h.Connections[uID] {
		conns = append(conns, conn)
	}
	return conns
}

func (h *Hub) Count(uID int64) int {
	h.Mu.RLock()
	defer h.Mu.RUnlock()
	return len(h.Connections[uID])
}
