package main

import (
	"log"
	"os"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/db"
	"github.com/charliekim2/chatapp/handler"
	"github.com/charliekim2/chatapp/lib"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {

	// TODO: refactor handlers into groups, function that sets up handlers
	// organize views into packages?
	app := pocketbase.New()

	hub := make(lib.Hub)

	app.OnModelAfterCreate("messages").Add(db.OnMessageEvent(hub, "create", app))
	app.OnModelAfterUpdate("messages").Add(db.OnMessageEvent(hub, "update", app))
	app.OnModelAfterDelete("messages").Add(db.OnMessageEvent(hub, "delete", app))

	app.OnModelAfterCreate("channels").Add(db.OnChannelCreate(app))

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(auth.LoadAuthContextFromCookie(app))

		// Static
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("static"), false))

		// Auth/signup
		e.Router.GET("/login", auth.GetLoginHandler)
		e.Router.POST("/login", auth.PostLoginHandler(app))
		e.Router.GET("/signup", handler.GetSignupHandler)
		e.Router.POST("/signup", handler.PostSignupHandler(app))

		//User operations
		e.Router.GET("/editprofile", handler.EditProfileHandler(app))
		e.Router.POST("/editprofile", handler.UpdateUserHandler(app))

		// UserChannel operations
		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(302, "/channels")
		})
		e.Router.POST("/subscribe", handler.SubscribeChannelHandler(app))
		e.Router.GET("/channels", handler.GetChannelsHandler(app))

		// Channel operations
		e.Router.POST("/channel", handler.CreateChannelHandler(app))
		e.Router.DELETE("/editchannel/:channelId", handler.DeleteChannelHandler(app))
		e.Router.POST("/editchannel/:channelId", handler.EditChannelHandler(app))

		// Chat/message operations
		e.Router.GET("/chat/:channel", handler.GetChatHandler(app))
		e.Router.GET("/livechat/:channel", handler.LiveChatHandler(app, hub))
		e.Router.GET("/messagechunk/:channel", handler.MessageChunkHandler(app))
		e.Router.PUT("/editmessage/:messageId", handler.EditMessageHandler(app))
		e.Router.DELETE("/editmessage/:messageId", handler.DeleteMessageHandler(app))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
