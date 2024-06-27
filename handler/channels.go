package handler

import (
	"net/http"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view/layout"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

func GetChannelsHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}
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

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Error getting channels for user")
		}

		return lib.Render(c, 200, layout.Channels(channels))
	}
}

func SubscribeChannelHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.FormValue("channelId")
		password := c.FormValue("password")

		channel, err := app.Dao().FindRecordById("channels", channelId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Channel not found")
		}

		if channel.Get("password") != password {
			return echo.NewHTTPError(http.StatusUnauthorized, "Incorrect password")
		}

		_, err = auth.AuthUserChannel(app, authRecord.Id, channelId)
		if err == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Already subscribed to channel")
		}

		usersChannels, err := app.Dao().FindCollectionByNameOrId("users_channels")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not subscribe to channel")
		}

		record := models.NewRecord(usersChannels)
		form := forms.NewRecordUpsert(app, record)
		form.LoadData(map[string]any{
			"userId":    authRecord.Id,
			"channelId": channelId,
		})

		if err = form.Submit(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not subscribe to channel")
		}

		return c.Redirect(http.StatusFound, "/chat/"+channelId)
	}
}

func CreateChannelHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		name := c.FormValue("channelName")
		password := c.FormValue("password")

		channels, err := app.Dao().FindCollectionByNameOrId("channels")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not create channel")
		}

		record := models.NewRecord(channels)
		form := forms.NewRecordUpsert(app, record)
		form.LoadData(map[string]any{
			"name":     name,
			"password": password,
			"ownerId":  authRecord.Id,
		})

		if err = form.Submit(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not create channel")
		}

		channelId := record.GetId()

		// TODO: make owner subcription to channel transactional/atomic
		usersChannels, err := app.Dao().FindCollectionByNameOrId("users_channels")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not subscribe to channel")
		}

		UCRecord := models.NewRecord(usersChannels)
		UCForm := forms.NewRecordUpsert(app, UCRecord)
		UCForm.LoadData(map[string]any{
			"userId":    authRecord.Id,
			"channelId": channelId,
		})

		if err = UCForm.Submit(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not subscribe to channel")
		}

		return c.Redirect(http.StatusFound, "/chat/"+channelId)
	}
}
