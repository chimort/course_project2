package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type MatchEvent struct {
	Type    string `json:"type"`
	ChatID  string `json:"chat_id"`
	Partner string `json:"partner"`
}

type WSHub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func NewWSHub() *WSHub {
	return &WSHub{
		conns: make(map[string]*websocket.Conn),
	}
}

func (h *WSHub) Set(username string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.conns[username]; ok && old != nil && old != conn {
		_ = old.Close()
	}

	h.conns[username] = conn
}

func (h *WSHub) Remove(username string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if c, ok := h.conns[username]; ok && c == conn {
		_ = c.Close()
		delete(h.conns, username)
	}
}

func (h *WSHub) Send(username string, payload any) error {

	h.mu.RLock()
	conn, ok := h.conns[username]
	h.mu.RUnlock()

	if !ok {
		return nil
	}

	return conn.WriteJSON(payload)
}
