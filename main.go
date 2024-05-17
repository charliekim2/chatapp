package main

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/charliekim2/chatapp/templates"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"
)

func loadAuthContextFromCookie(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenCookie, err := c.Request().Cookie("token")
			if err != nil || tokenCookie.Value == "" {
				return next(c) // no token cookie
			}

			token := tokenCookie.Value

			claims, _ := security.ParseUnverifiedJWT(token)
			tokenType := cast.ToString(claims["type"])

			switch tokenType {
			case tokens.TypeAdmin:
				admin, err := app.Dao().FindAdminByToken(
					token,
					app.Settings().AdminAuthToken.Secret,
				)
				if err == nil && admin != nil {
					// "authenticate" the admin
					c.Set(apis.ContextAdminKey, admin)
				}
			case tokens.TypeAuthRecord:
				record, err := app.Dao().FindAuthRecordByToken(
					token,
					app.Settings().RecordAuthToken.Secret,
				)
				if err == nil && record != nil {
					// "authenticate" the app user
					c.Set(apis.ContextAuthRecordKey, record)
				}
			}

			return next(c)
		}
	}
}

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
        e.Router.Use(loadAuthContextFromCookie(app))

        e.Router.GET("/hello/:name", helloHandler)
        e.Router.GET("/login", loginHandler)
        return nil
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}