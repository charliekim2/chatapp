package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

var (
	upgrader = websocket.Upgrader{}
)

func GetChatHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		channel, err := AuthUserChannel(app, authRecord.Id, channelId)
		if err != nil {
			return err
		}

		messages := []model.Message{}

		// TODO: message model contains owner name, profile picture path, etc.
		err = app.Dao().DB().
			NewQuery(
				"SELECT id, ownerId, created, body " +
					"FROM messages " +
					"WHERE channelId = {:channelId}" +
					"ORDER BY MESSAGES.created ASC " +
					"LIMIT 50;",
			).
			Bind(dbx.Params{"channelId": channelId}).
			All(&messages)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Could not find messages in channel")
		}

		return lib.Render(c, 200, view.Chat(messages, &channel))
	}
}

func LiveChatHandler(app *pocketbase.PocketBase, hub Hub) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		_, err := AuthUserChannel(app, authRecord.Id, channelId)
		if err != nil {
			return err
		}

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not init websocket")
		}

		chat, ok := hub[channelId]

		if !ok {
			hub[channelId] = &Chat{
				id:         channelId,
				clients:    make(map[*Client]bool),
				broadcast:  make(chan []byte),
				register:   make(chan *Client),
				unregister: make(chan *Client),
			}

			chat = hub[channelId]
			go chat.run(app)
		}

		client := &Client{
			id:   authRecord.Id,
			chat: chat,
			conn: ws,
			send: make(chan []byte, 512),
		}

		client.chat.register <- client

		go client.writePump()
		go client.readPump()

		return nil
	}
}

func AuthUserChannel(app *pocketbase.PocketBase, userId string, channelId string) (model.Channel, error) {
	channel := model.Channel{}

	err := app.Dao().DB().
		NewQuery(
			"SELECT CHANNELS.name, CHANNELS.id " +
				"FROM CHANNELS " +
				"JOIN USERS_CHANNELS ON CHANNELS.id = USERS_CHANNELS.channelId " +
				"WHERE USERS_CHANNELS.userId = {:userId} AND USERS_CHANNELS.channelId = {:channelId};",
		).
		Bind(dbx.Params{"userId": userId, "channelId": channelId}).
		One(&channel)

	if err != nil {
		return model.Channel{}, echo.NewHTTPError(http.StatusNotFound, "Could not connect to channel")
	}

	return channel, nil
}

// Credit: https://github.com/gorilla/websocket/tree/main/examples/chat
// A user connection to a live chat
type Client struct {
	id   string
	chat *Chat
	conn *websocket.Conn
	send chan []byte
}

// A live chat with a map of connected clients
type Chat struct {
	id         string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func (c *Chat) GetClients() map[*Client]bool {
	return c.clients
}

func (c *Client) GetSend() chan []byte {
	return c.send
}

// Channel IDs mapped to open chats
type Hub map[string]*Chat

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

// TODO: finish these
func (c *Client) writePump() {
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

func (c *Client) readPump() {
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

		msgObject.OwnerId = c.id
		message, err = json.Marshal(msgObject)
		if err != nil {
			// TODO: some sort of client alert that message could not be sent? -> send a template that targets a notification element
			log.Println("Could not marshal the message")
			continue
		}

		c.chat.broadcast <- message
	}
}

func (c *Chat) run(app *pocketbase.PocketBase) {
	for {
		select {
		case client := <-c.register:
			c.clients[client] = true
		case client := <-c.unregister:
			// TODO: once all clients unregister, delete chat from hub
			if _, ok := c.clients[client]; ok {
				delete(c.clients, client)
				close(client.send)
			}
		case message := <-c.broadcast:
			// json parse message value
			// insert into database
			// add listener for messages db, get channelID
			// if channel is in hub, parse message into template and send to clients
			var msgObject model.DBMessage
			err := json.Unmarshal(message, &msgObject)
			if err != nil {
				log.Println("Could not unmarshal the message")
				continue
			}

			msgObject.ChannelId = c.id
			collection, err := app.Dao().FindCollectionByNameOrId("messages")
			// TODO: some sort of client alert that message could not be sent? -> send a template that htmx-targets a notification element
			if err != nil {
				log.Println("Could not find messages table")
				continue
			}
			// TODO: validate/cleanse data
			record := models.NewRecord(collection)

			record.Set("body", msgObject.Body)
			record.Set("ownerId", msgObject.OwnerId)
			record.Set("channelId", msgObject.ChannelId)

			if err = app.Dao().SaveRecord(record); err != nil {
				log.Println("Could not save message to db")
				continue
			}
		}
	}
}
