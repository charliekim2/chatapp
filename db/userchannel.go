package db

import (
	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
)

func CreateUserChannel(app *pocketbase.PocketBase, userChannel *model.UserChannel) error {
	return nil
}

func ReadUserSubscriptions(app *pocketbase.PocketBase, userChannel *model.UserChannel) error {
	return nil
}

func ReadChannelSubscribers(app *pocketbase.PocketBase, userChannel *model.UserChannel) error {
	return nil
}

func DeleteUserChannel(app *pocketbase.PocketBase, userChannel *model.UserChannel) error {
	return nil
}
