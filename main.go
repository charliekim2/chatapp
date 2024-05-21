package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"
)

// TODO: refactor all of this into packages

func loadAuthContextFromCookie(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenCookie, err := c.Request().Cookie("pb_auth")
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

	return Render(c, 200, view.Hello(name))
}

func getLoginHandler(c echo.Context) error {
	return Render(c, 200, view.Login())
}

func postLoginHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		users, err := app.Dao().FindCollectionByNameOrId("users")

		if err != nil {
			// todo: custom error page
			return echo.NewHTTPError(http.StatusNotFound, "Error querying users")
		}

		form := forms.NewRecordPasswordLogin(app, users)
		c.Bind(form)
		authRecord, err := form.Submit()

		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Error validating login")
		}

		token, err := tokens.NewRecordAuthToken(app, authRecord)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Error generating login token")
		}

		c.SetCookie(&http.Cookie{
			Name:     "pb_auth",
			Value:    token,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			HttpOnly: true,
			MaxAge:   int(app.Settings().RecordAuthToken.Duration),
			Path:     "/",
		})

		return c.Redirect(http.StatusFound, "/")
	}
}

func getChannelsHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channels := []model.Channel{}

		err := app.Dao().DB().
			NewQuery(
				"SELECT CHANNELS.name, CHANNELS.id " +
					"FROM CHANNELS " +
					"JOIN USERS_CHANNELS ON CHANNELS.id = USERS_CHANNELS.channelId " +
					"WHERE USERS_CHANNELS.userId = {:userId};",
			).
			Bind(dbx.Params{"userId": authRecord.Id}).
			All(&channels)

		fmt.Println(channels)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Error getting channels for user")
		}

		return Render(c, 200, view.Channels(channels))
	}
}

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(loadAuthContextFromCookie(app))

		e.Router.GET("/hello/:name", helloHandler)
		e.Router.GET("/login", getLoginHandler)
		e.Router.POST("/login", postLoginHandler(app))
		e.Router.GET("/", getChannelsHandler(app))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
