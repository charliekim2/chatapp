package handler

import (
	"log"
	"net/http"

	"github.com/charliekim2/chatapp/db"
	"github.com/charliekim2/chatapp/model"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

func DeleteMessageHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		messageId := c.PathParam("messageId")
		msg, err := db.ReadMessage(app, messageId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Message not found")
		}
		if msg.OwnerId != authRecord.Id {
			return echo.NewHTTPError(http.StatusUnauthorized, "You do not have permission to delete this message")
		}

		err = db.DeleteMessage(app, messageId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error deleting message")
		}
		return c.NoContent(http.StatusNoContent)
	}
}

func EditMessageHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		messageId := c.PathParam("messageId")
		msg, err := db.ReadMessage(app, messageId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Message not found")
		}
		if msg.OwnerId != authRecord.Id {
			return echo.NewHTTPError(http.StatusUnauthorized, "You do not have permission to edit this message")
		}

		body := c.FormValue("body")
		updatedMessage := &model.DBMessage{
			Id:   messageId,
			Body: body,
		}
		err = db.UpdateMessage(app, updatedMessage)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error updating message")
		}

		// Return 204 so HTMX doesn't swap the form content; let websocket swap the message instead
		return c.NoContent(http.StatusNoContent)
	}
}
