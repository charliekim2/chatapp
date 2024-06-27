package db

import (
	"bytes"
	"context"

	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view/component"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

func OnMessageCreate(hub lib.Hub) func(e *core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		channelId := e.Model.(*models.Record).Get("channelId")
		chat, ok := hub[channelId.(string)]
		if !ok {
			return nil
		}

		id := e.Model.GetId()
		body := e.Model.(*models.Record).Get("body").(string)
		// TODO: createdAt should be Time and let the templates do the conversion to string
		createdAt := e.Model.(*models.Record).Get("created").(types.DateTime).String()
		ownerId := e.Model.(*models.Record).Get("ownerId").(string)

		msg := model.Message{Id: id, Body: body, CreatedAt: createdAt, OwnerId: ownerId}
		component := component.Message(&msg)
		var b []byte
		buf := bytes.NewBuffer(b)
		component.Render(context.Background(), buf)

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
