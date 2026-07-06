// client.go — 单个 WebSocket 连接的封装与读写循环。
//
// 声明内容：
//   - 连接相关常量：writeWait、pongWait、pingPeriod、maxMessageSize，
//     以及 RFC 6455 消息类型（pingMessage、closeMessage，GoFr 封装未暴露）
//   - Client：封装单个 WebSocket 连接，持有 hub、conn、send 缓冲通道和 userID
//
// 职责：
//   - ReadPump：从连接读取客户端消息并交给 MessageHandler 处理，设置读超时并通过 pong 重置
//   - WritePump：从 send 通道取出消息写入连接，并定期发送 ping 维持心跳
//   - 提供 NewClient / NewClientFromConn 构造方法与 Send 消息投递方法
package ws

import (
	"time"

	gofrWS "gofr.dev/pkg/gofr/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096

	// RFC 6455 message types not exposed by GoFr's websocket wrapper.
	pingMessage  = 9
	closeMessage = 8
)

// Client wraps a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *gofrWS.Connection
	send   chan []byte
	userID int64
}

// NewClient creates a new Client.
func NewClient(hub *Hub, conn *gofrWS.Connection, userID int64) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}
}

// NewClientFromConn creates a new Client from a GoFr websocket connection.
func NewClientFromConn(hub *Hub, conn *gofrWS.Connection, userID int64) *Client {
	return NewClient(hub, conn, userID)
}

// Send queues a message to be written to the WebSocket connection.
func (c *Client) Send(data []byte) {
	c.send <- data
}

// ReadPump reads messages from the WebSocket connection and dispatches them.
func (c *Client) ReadPump(handler MessageHandler) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		handler.Handle(c, message)
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(closeMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(gofrWS.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(pingMessage, nil); err != nil {
				return
			}
		}
	}
}
