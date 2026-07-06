// hub.go — WebSocket 连接与会话路由管理中心。
//
// 声明内容：
//   - WSMessage：推送给客户端的消息结构（Type + Data）
//   - Hub：管理所有 WebSocket 连接和会话成员路由的核心结构
//   - broadcastMsg：广播消息的内部载体（sessionID + message）
//
// 职责：
//   - 维护在线客户端连接表（clients，按 userID 分组）和会话成员表（sessions，按 sessionID 分组）
//   - 通过 register/unregister/broadcast 三个 channel 驱动单 goroutine 事件循环（Run），
//     将对连接表的修改串行化，降低锁争用
//   - 对外提供按会话广播（BroadcastToSession）、按用户发送（SendToUser）、
//     会话订阅/退订（RegisterSession/UnregisterSession）以及在线状态查询（IsOnline）等能力
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
	// clients 按 userID 分组保存所有在线 WebSocket 连接。
	// 外层 key 为 userID，内层 map 为该用户的所有客户端连接（支持单用户多端同时在线）。
	clients map[int64]map[*Client]bool
	// sessions 按 sessionID 分组保存会话成员的 userID。
	// 外层 key 为会话 ID（单聊或群聊），内层 map 为该会话的所有成员 userID，用于广播消息路由。
	sessions map[int64]map[int64]bool

	// register 注册通道，新客户端建立连接时通过该通道提交给 Hub 主循环处理。
	register chan *Client
	// unregister 注销通道，客户端断开连接时通过该通道提交给 Hub 主循环清理资源。
	unregister chan *Client
	// broadcast 广播通道，向指定会话的所有在线成员推送消息时使用。
	broadcast chan *broadcastMsg

	// mu 读写锁，保护 clients 与 sessions 两个 map 的并发安全访问。
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
