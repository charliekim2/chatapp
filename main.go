package main

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/charliekim2/chatapp/templates"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func Render(c echo.Context, status int, t templ.Component) error {
    c.Response().Writer.WriteHeader(status)

    err := t.Render(context.Background(), c.Response().Writer)

    if err != nil {
        return c.String(http.StatusInternalServerError, "Error rendering template")
    }

    return nil
}

func helloHandler(c echo.Context) error {
    name := c.PathParam("name")

    return Render(c, 200, templates.Hello(name))
}

func loginHandler(c echo.Context) error {
    return Render(c, 200, templates.Login())
}


func main() {
    app := pocketbase.New()

    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        e.Router.GET("/hello/:name", helloHandler)
        e.Router.GET("/login", loginHandler)
        return nil
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}