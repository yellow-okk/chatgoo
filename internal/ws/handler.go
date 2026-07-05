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
