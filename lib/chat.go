package lib

import (
	"encoding/json"
	"log"

	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

const CHUNK = 10

// Credit: https://github.com/gorilla/websocket/tree/main/examples/chat
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

func (c *Chat) GetRegister() chan *Client {
	return c.register
}

func NewChat(id string) *Chat {
	return &Chat{
		id:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Channel IDs mapped to open chats
type Hub map[string]*Chat

func (c *Chat) Run(app *pocketbase.PocketBase) {
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
