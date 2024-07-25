package lib

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/charliekim2/chatapp/model"
	"github.com/gorilla/websocket"
)

// Credit: https://github.com/gorilla/websocket/tree/main/examples/chat
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// A user connection to a live chat
type Client struct {
	user *model.User
	chat *Chat
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) GetSend() chan []byte {
	return c.send
}

func (c *Client) GetChat() *Chat {
	return c.chat
}

func (c *Client) GetId() string {
	return c.user.Id
}

func (c *Client) GetUser() *model.User {
	return c.user
}

func NewClient(user *model.User, chat *Chat, conn *websocket.Conn) *Client {
	return &Client{
		user: user,
		chat: chat,
		conn: conn,
		send: make(chan []byte, 512),
	}
}

// TODO: finish these
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.chat.unregister <- c
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// add client ID to message JSON
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var msgObject model.DBMessage
		err = json.Unmarshal(message, &msgObject)
		if err != nil {
			// TODO: some sort of client alert that message could not be sent? -> send a template that targets a notification element
			log.Println("Could not unmarshal the message")
			continue
		}

		msgObject.OwnerId = c.GetId()
		message, err = json.Marshal(msgObject)
		if err != nil {
			// TODO: some sort of client alert that message could not be sent? -> send a template that targets a notification element
			log.Println("Could not marshal the message")
			continue
		}

		c.chat.broadcast <- message
	}
}
