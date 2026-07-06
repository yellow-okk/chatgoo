// handler.go — WebSocket 入站消息的处理器。
//
// 声明内容：
//   - MessageHandler 接口：定义处理客户端消息的契约（Handle）
//   - defaultMessageHandler：默认实现，持有 Hub 引用
//   - incomingMsg / subscribeData：入站 JSON 消息的解析结构
//
// 职责：
//   - 解析客户端发来的 JSON 消息，按 type 字段分发到对应处理逻辑
//   - 支持：ping 心跳响应（回 pong）、subscribe_session 订阅会话、
//     unsubscribe_session 退订会话
package ws

import (
	"encoding/json"
)

// MessageHandler processes incoming WebSocket messages.
type MessageHandler interface {
	Handle(client *Client, message []byte)
}

type defaultMessageHandler struct {
	hub *Hub
}

// NewDefaultMessageHandler creates the default message handler.
func NewDefaultMessageHandler(hub *Hub) MessageHandler {
	return &defaultMessageHandler{hub: hub}
}

type incomingMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type subscribeData struct {
	SessionID int64 `json:"session_id"`
}

// Handle processes incoming WebSocket messages.
func (h *defaultMessageHandler) Handle(client *Client, message []byte) {
	var msg incomingMsg
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "ping":
		resp, _ := json.Marshal(WSMessage{Type: "pong", Data: nil})
		client.send <- resp

	case "subscribe_session":
		var data subscribeData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return
		}
		h.hub.RegisterSession(data.SessionID, client.userID)

	case "unsubscribe_session":
		var data subscribeData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return
		}
		h.hub.UnregisterSession(data.SessionID, client.userID)
	}
}
