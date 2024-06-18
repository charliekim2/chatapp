package auth

import (
	"net/http"

	"github.com/charliekim2/chatapp/model"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"
)

// Credit: https://github.com/pocketbase/pocketbase/discussions/989
func LoadAuthContextFromCookie(app core.App) echo.MiddlewareFunc {
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
