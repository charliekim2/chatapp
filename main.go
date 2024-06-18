package main

import (
	"bytes"
	"context"
	"log"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/handler"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

func helloHandler(c echo.Context) error {
	name := c.PathParam("name")

	return lib.Render(c, 200, view.Hello(name))
}

func main() {

	//TODO: refactor handlers into groups, function that sets up handlers
	// organize views into packages?
	app := pocketbase.New()

	hub := make(handler.Hub)

	app.OnModelAfterCreate("messages").Add(func(e *core.ModelEvent) error {
		channelId := e.Model.(*models.Record).Get("channelId")
		chat, ok := hub[channelId.(string)]
		if !ok {
			return nil
		}

		id := e.Model.GetId()
		body := e.Model.(*models.Record).Get("body").(string)
		// TODO: createdAt should be Time and let the templates do the conversion to string
		createdAt := e.Model.(*models.Record).Get("created").(types.DateTime).String()
		ownerId := e.Model.(*models.Record).Get("ownerId").(string)

		msg := model.Message{Id: id, Body: body, CreatedAt: createdAt, OwnerId: ownerId}
		component := view.Message(&msg)
		var b []byte
		buf := bytes.NewBuffer(b)
		component.Render(context.Background(), buf)

		for client := range chat.GetClients() {
			select {
			case client.GetSend() <- buf.Bytes():
			default:
				close(client.GetSend())
				delete(chat.GetClients(), client)
			}
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(auth.LoadAuthContextFromCookie(app))

		e.Router.GET("/hello/:name", helloHandler)
		e.Router.GET("/login", auth.GetLoginHandler)
		e.Router.POST("/login", auth.PostLoginHandler(app))
		e.Router.GET("/", handler.GetChannelsHandler(app))
		e.Router.GET("/chat/:channel", handler.GetChatHandler(app))
		e.Router.GET("/livechat/:channel", handler.LiveChatHandler(app, hub))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
