package main

import (
	"log"
	"os"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/db"
	"github.com/charliekim2/chatapp/handler"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func helloHandler(c echo.Context) error {
	name := c.PathParam("name")

	return lib.Render(c, 200, view.Hello(name))
}

func main() {

	//TODO: refactor handlers into groups, function that sets up handlers
	// organize views into packages?
	app := pocketbase.New()

	hub := make(lib.Hub)

	app.OnModelAfterCreate("messages").Add(db.OnMessageCreate(hub))

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(auth.LoadAuthContextFromCookie(app))

		e.Router.GET("/hello/:name", helloHandler)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("static"), false))

		e.Router.GET("/login", auth.GetLoginHandler)
		e.Router.POST("/login", auth.PostLoginHandler(app))
		e.Router.GET("/signup", handler.GetSignupHandler)
		e.Router.POST("/signup", handler.PostSignupHandler(app))

		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(302, "/channels")
		})
		e.Router.GET("/channels", handler.GetChannelsHandler(app))
		e.Router.POST("/subscribe", handler.SubscribeChannelHandler(app))
		e.Router.GET("/chat/:channel", handler.GetChatHandler(app))
		e.Router.GET("/livechat/:channel", handler.LiveChatHandler(app, hub))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
