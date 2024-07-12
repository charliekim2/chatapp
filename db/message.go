package db

import (
	"bytes"
	"context"
	"errors"

	"github.com/a-h/templ"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view/component"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

func OnMessageEvent(hub lib.Hub, eventType string) func(e *core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		channelId := e.Model.(*models.Record).Get("channelId")
		chat, ok := hub[channelId.(string)]
		if !ok {
			return nil
		}

		id := e.Model.GetId()
		body := e.Model.(*models.Record).Get("body").(string)
		createdAt := e.Model.(*models.Record).Get("created").(types.DateTime).String()
		ownerId := e.Model.(*models.Record).Get("ownerId").(string)

		msg := model.Message{Id: id, Body: body, CreatedAt: createdAt, OwnerId: ownerId}

		var cmpt templ.Component
		switch eventType {
		case "create":
			cmpt = component.Message(&msg)
		case "update":
			cmpt = component.EditMessage(&msg)
		case "delete":
			cmpt = component.DeleteMessage(&msg)
		default:
			return errors.New("Invalid event type")
		}

		var b []byte
		buf := bytes.NewBuffer(b)
		cmpt.Render(context.Background(), buf)

		for client := range chat.GetClients() {
			select {
			case client.GetSend() <- buf.Bytes():
			default:
				close(client.GetSend())
				delete(chat.GetClients(), client)
			}
		}
		return nil
	}
}

func CreateMessage(app *pocketbase.PocketBase, message *model.DBMessage) (string, error) {
	messages, err := app.Dao().FindCollectionByNameOrId("messages")
	if err != nil {
		return "", err
	}

	record := models.NewRecord(messages)
	form := forms.NewRecordUpsert(app, record)
	form.LoadData(map[string]interface{}{
		"body":      message.Body,
		"ownerId":   message.OwnerId,
		"channelId": message.ChannelId,
	})

	if err = form.Submit(); err != nil {
		return "", err
	}
	messageId := record.GetId()
	return messageId, nil
}

func ReadMessage(app *pocketbase.PocketBase, messageId string) (*model.Message, error) {
	record, err := app.Dao().FindRecordById("messages", messageId)
	if err != nil {
		return nil, err
	}

	message := &model.Message{
		Id:        record.GetId(),
		Body:      record.Get("body").(string),
		CreatedAt: record.Get("created").(types.DateTime).String(),
		OwnerId:   record.Get("ownerId").(string),
		ChannelId: record.Get("channelId").(string),
	}
	return message, nil
}

func UpdateMessage(app *pocketbase.PocketBase, message *model.DBMessage) error {
	record, err := app.Dao().FindRecordById("messages", message.Id)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app, record)
	form.LoadData(map[string]interface{}{
		"body": message.Body,
	})

	if err = form.Submit(); err != nil {
		return err
	}
	return nil
}

func DeleteMessage(app *pocketbase.PocketBase, messageId string) error {
	record, err := app.Dao().FindRecordById("messages", messageId)
	if err != nil {
		return err
	}

	if err = app.Dao().DeleteRecord(record); err != nil {
		return err
	}
	return nil
}
