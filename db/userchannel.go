package db

import (
	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
)

// TODO: event hook for unsubscribe - if the user was owner, delete channel/tranfer ownership

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

func SubscribeChannel() {}

func UnsubscribeChannel() {}
