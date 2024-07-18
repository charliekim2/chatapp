package db

import (
	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
)

func ReadUser(app *pocketbase.PocketBase, userId string) (*model.User, error) {
	record, err := app.Dao().FindRecordById("users", userId)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Id:       record.GetId(),
		Username: record.Get("username").(string),
		Name:     record.Get("name").(string),
	}
	return user, nil
}
