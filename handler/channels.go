package handler

import (
	"fmt"
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

func GetChannelsHandler(app *pocketbase.PocketBase) func(echo.Context) error {
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

		return lib.Render(c, 200, view.Channels(channels))
	}
}
