package handler

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/charliekim2/chatapp/auth"
	"github.com/charliekim2/chatapp/db"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view/component"
	"github.com/charliekim2/chatapp/view/layout"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

var (
	upgrader = websocket.Upgrader{}
)

func GetChatHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		channel, err := auth.AuthUserChannel(app, authRecord.Id, channelId)
		if err != nil {
			return err
		}

		messages := []model.Message{}

		// TODO: message model contains owner name, profile picture path, etc.
		err = app.Dao().DB().
			Select("id", "ownerId", "created", "body").
			From("messages").
			Where(dbx.NewExp("channelId = {:channelId}", dbx.Params{"channelId": channelId})).
			OrderBy("MESSAGES.created DESC").
			Limit(lib.CHUNK).
			All(&messages)

		slices.Reverse(messages)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Could not find messages in channel")
		}

		user, err := db.ReadUser(app, authRecord.Id)
		return lib.Render(c, 200, layout.Chat(messages, &channel, user))
	}
}

func LiveChatHandler(app *pocketbase.PocketBase, hub lib.Hub) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		_, err := auth.AuthUserChannel(app, authRecord.Id, channelId)
		if err != nil {
			return err
		}

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not init websocket")
		}

		chat, ok := hub[channelId]

		if !ok {
			hub[channelId] = lib.NewChat(channelId)

			chat = hub[channelId]
			go chat.Run(app)
		}

		user, err := db.ReadUser(app, authRecord.Id)
		client := lib.NewClient(user, chat, ws)

		client.GetChat().GetRegister() <- client

		go client.WritePump()
		go client.ReadPump()

		return nil
	}
}

func MessageChunkHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		channelId := c.PathParam("channel")
		channel, err := auth.AuthUserChannel(app, authRecord.Id, channelId)
		if err != nil {
			return err
		}

		messages := []model.Message{}
		offset, err := strconv.Atoi(c.QueryParam("offset"))

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset")
		}

		err = app.Dao().DB().
			Select("id", "ownerId", "created", "body").
			From("messages").
			Where(dbx.NewExp("channelId = {:channelId}", dbx.Params{"channelId": channelId})).
			OrderBy("MESSAGES.created DESC").
			Limit(lib.CHUNK).
			Offset(int64(offset)).
			Bind(dbx.Params{"channelId": channelId}).
			All(&messages)

		if len(messages) == 0 {
			return nil
		}

		slices.Reverse(messages)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Could not find messages in channel")
		}

		user, err := db.ReadUser(app, authRecord.Id)
		return lib.Render(c, 200, component.MessageChunk(messages, user, &channel, offset))
	}
}
