package handler

import (
	"encoding/json"
	"sync"

	"chatgoo/internal/pkg/jwt"
	"chatgoo/internal/ws"

	gofrWS "gofr.dev/pkg/gofr/websocket"
	"gofr.dev/pkg/gofr"
)

// WSHandler returns a GoFr WebSocket handler that integrates with the Hub.
// Auth is handled via ?token= query parameter since app.WebSocket doesn't support per-route middleware.
func WSHandler(hub *ws.Hub, jwtSecret string) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		token := c.Param("token")
		if token == "" {
			return nil, gofrWS.ErrorConnection
		}

		claims, err := jwt.Parse(token, jwtSecret)
		if err != nil {
			return nil, gofrWS.ErrorConnection
		}
		userID := claims.UserID

		conn := c.Container.GetConnectionFromContext(c.Context)
		if conn == nil || conn.Conn == nil {
			return nil, gofrWS.ErrorConnection
		}

		client := ws.NewClientFromConn(hub, conn.Conn, userID)
		hub.RegisterClient(client)

		var once sync.Once
		cleanup := func() {
			once.Do(func() {
				hub.UnregisterClient(client)
			})
		}
		defer cleanup()

		go func() {
			client.WritePump()
			cleanup()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return nil, err
			}

			var msg struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "ping":
				resp, _ := json.Marshal(ws.WSMessage{Type: "pong", Data: nil})
				client.Send(resp)

			case "subscribe_session":
				var data struct {
					SessionID int64 `json:"session_id"`
				}
				if json.Unmarshal(msg.Data, &data) == nil {
					hub.RegisterSession(data.SessionID, userID)
				}

			case "unsubscribe_session":
				var data struct {
					SessionID int64 `json:"session_id"`
				}
				if json.Unmarshal(msg.Data, &data) == nil {
					hub.UnregisterSession(data.SessionID, userID)
				}
			}
		}
	}
}
