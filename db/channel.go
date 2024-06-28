package db

import (
	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

// Subscribe owner to newly created channel
func OnChannelCreate(app *pocketbase.PocketBase) func(e *core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		channelId := e.Model.GetId()
		ownerId := e.Model.(*models.Record).Get("ownerId").(string)

		usersChannels, err := app.Dao().FindCollectionByNameOrId("users_channels")
		if err != nil {
			return err
		}

		record := models.NewRecord(usersChannels)
		form := forms.NewRecordUpsert(app, record)
		form.LoadData(map[string]any{
			"userId":    ownerId,
			"channelId": channelId,
		})

		if err = form.Submit(); err != nil {
			return err
		}
		return nil
	}
}

func CreateChannel(app *pocketbase.PocketBase, channel *model.DBChannel) (string, error) {
	channels, err := app.Dao().FindCollectionByNameOrId("channels")
	if err != nil {
		return "", err
	}

	record := models.NewRecord(channels)
	form := forms.NewRecordUpsert(app, record)
	form.LoadData(map[string]any{
		"name":     channel.Name,
		"password": channel.Password,
		"ownerId":  channel.OwnerId,
	})

	if err = form.Submit(); err != nil {
		return "", err
	}
	channelId := record.GetId()
	return channelId, nil
}

func ReadChannel(app *pocketbase.PocketBase, channelId string) (*model.Channel, error) {
	record, err := app.Dao().FindRecordById("channels", channelId)
	if err != nil {
		return nil, err
	}

	channel := &model.Channel{
		Id:   record.Id,
		Name: record.Get("name").(string),
	}

	return channel, nil
}

func UpdateChannel(app *pocketbase.PocketBase, channel *model.DBChannel) error {
	record, err := app.Dao().FindRecordById("channels", channel.Id)
	if err != nil {
		return err
	}

	cmap := map[string]any{}
	if channel.Name != "" {
		cmap["name"] = channel.Name
	}
	if channel.OwnerId != "" {
		cmap["ownerId"] = channel.OwnerId
	}
	if channel.Password != "" {
		cmap["password"] = channel.Password
	}

	form := forms.NewRecordUpsert(app, record)
	form.LoadData(cmap)

	if err = form.Submit(); err != nil {
		return err
	}
	return nil
}

func DeleteChannel(app *pocketbase.PocketBase, channelId string) error {
	record, err := app.Dao().FindRecordById("channels", channelId)
	if err != nil {
		return err
	}

	if err = app.Dao().DeleteRecord(record); err != nil {
		return err
	}
	return nil
}
