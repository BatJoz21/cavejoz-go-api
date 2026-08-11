package hub

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan any
}

func NewClient(uID int64, conn *websocket.Conn) *Client {
	return &Client{
		UserID: uID,
		Conn:   conn,
		Send:   make(chan any, 256),
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.Send:
			if !ok {
				// hub closed the channel — connection is finished
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteJSON(payload); err != nil {
				return
			}

		case <-ticker.C:
			log.Printf("ping -> user %d", c.UserID)
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump(h *Hub) {
	defer func() {
		h.Remove(c.UserID, c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("ping <- user %d", c.UserID)
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				log.Printf("User %d dropped unexpectedly: %v", c.UserID, err)
			}
			break
		}
	}
}
