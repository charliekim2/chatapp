package db

import (
	"github.com/charliekim2/chatapp/model"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
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

func UpdateUser(app *pocketbase.PocketBase, user *model.DBUser) error {
	record, err := app.Dao().FindRecordById("users", user.Id)
	if err != nil {
		return err
	}

	cmap := map[string]any{}
	if user.Name != "" {
		cmap["name"] = user.Name
		cmap["username"] = user.Name
	}
	if user.Password != "" {
		cmap["password"] = user.Password
		cmap["passwordConfirm"] = user.PasswordConfirm
		cmap["oldPassword"] = user.OldPassword
	}

	form := forms.NewRecordUpsert(app, record)
	form.LoadData(cmap)

	if err = form.Submit(); err != nil {
		return err
	}

	return nil
}
