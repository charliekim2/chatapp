package handler

import (
	"bytes"
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
		err := AuthUserChannel(app, authRecord.Id, channelId)
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

		return lib.Render(c, 200, view.Chat(messages))
	}
}

func LiveChatHandler(app *pocketbase.PocketBase, hub Hub) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		err := AuthUserChannel(app, authRecord.Id, channelId)
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

		// request websocket endpoint
		// init Client connection in Chat
		// keys: the pointer address?? ig
		// read and write endpoint, with mutexes ig
		// user sends message:
		// websocket reads, validates?, sends to database
		// on database update: listener that gets new message
		// render into <li> and send to get htmx swapped beforeend
		// determine which Client sent it -> get the Chat -> send to all Clients in that Chat
		// on websocket close -> delete the Client/close connection properly -> if Chat is empty, delete Chat from Hub
		return nil
	}
}

func AuthUserChannel(app *pocketbase.PocketBase, userId string, channelId string) error {
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
		return echo.NewHTTPError(http.StatusNotFound, "Could not connect to channel")
	}

	return nil
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
	defer func() {
		c.conn.Close()
	}()
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
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.chat.broadcast <- message
	}
}
