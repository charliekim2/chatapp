package handler

import (
	"net/http"

	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

// TODO: instead of closures, wrapper class for app that we define handler methods on?
func GetChatHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		channelId := c.PathParam("channel")
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channel := model.Channel{}

		// Determine if user is in channel requested, if not will it throw an error? Will have to test
		err := app.Dao().DB().
			NewQuery(
				"SELECT CHANNELS.name, CHANNELS.id " +
					"FROM CHANNELS " +
					"JOIN USERS_CHANNELS ON CHANNELS.id = USERS_CHANNELS.channelId " +
					"WHERE USERS_CHANNELS.userId = {:userId} AND USERS_CHANNELS.channelId = {:channelId};",
			).
			Bind(dbx.Params{"userId": authRecord.Id, "channelId": channelId}).
			One(&channel)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Could not find channel")
		}

		messages := []model.Message{}

		// TODO: message model contains owner name, profile picture path, etc.
		err = app.Dao().DB().
			NewQuery(
				"SELECT MESSAGES.id, MESSAGES.ownerId, MESSAGES.createdAt, MESSAGES.body" +
					"FROM MESSAGES " +
					"WHERE MESSAGES.channelId = {:channelId} " +
					"ORDER BY MESSAGES.createdAt ASC " +
					"LIMIT 50;",
			).
			Bind(dbx.Params{"channelId": channelId}).
			All(&messages)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Could not find messages in channel")
		}

		return lib.Render(c, 200, view.Chat())
	}
}
