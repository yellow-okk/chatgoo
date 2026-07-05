package ws

import (
	"encoding/json"
	"sync"
)

// WSMessage is the structure for WebSocket push messages.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub manages all WebSocket connections and session routing.
type Hub struct {
	clients  map[int64]map[*Client]bool
	sessions map[int64]map[int64]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan *broadcastMsg

	mu sync.RWMutex
}

type broadcastMsg struct {
	sessionID int64
	message   *WSMessage
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		sessions:   make(map[int64]map[int64]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *broadcastMsg, 256),
	}
}

// Run starts the hub's main event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.userID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			userIDs := h.sessions[msg.sessionID]
			h.mu.RUnlock()

			payload, _ := json.Marshal(msg.message)
			for uid := range userIDs {
				h.sendToUser(uid, payload)
			}
		}
	}
}

// RegisterSession registers a user to a session for broadcast routing.
func (h *Hub) RegisterSession(sessionID, userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[sessionID] == nil {
		h.sessions[sessionID] = make(map[int64]bool)
	}
	h.sessions[sessionID][userID] = true
}

// UnregisterSession removes a user from a session.
func (h *Hub) UnregisterSession(sessionID, userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if users, ok := h.sessions[sessionID]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(h.sessions, sessionID)
		}
	}
}

// BroadcastToSession sends a message to all online members of a session.
func (h *Hub) BroadcastToSession(sessionID int64, msg *WSMessage) {
	h.broadcast <- &broadcastMsg{sessionID: sessionID, message: msg}
}

// SendToUser sends a message to all connections of a specific user.
func (h *Hub) SendToUser(userID int64, msg *WSMessage) {
	payload, _ := json.Marshal(msg)
	h.sendToUser(userID, payload)
}

func (h *Hub) sendToUser(userID int64, payload []byte) {
	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for client := range conns {
		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}
}

// RegisterClient registers a new client with the hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client from the hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// IsOnline checks if a user has any active connections.
func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userID]
	return ok && len(conns) > 0
}
