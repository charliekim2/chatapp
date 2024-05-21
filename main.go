package main

import (
	"log"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/handler"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(auth.LoadAuthContextFromCookie(app))

		e.Router.GET("/hello/:name", helloHandler)
		e.Router.GET("/login", auth.GetLoginHandler)
		e.Router.POST("/login", auth.PostLoginHandler(app))
		e.Router.GET("/", handler.GetChannelsHandler(app))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
