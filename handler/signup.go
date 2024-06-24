package handler

import (
	"log"
	"net/http"

	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

type User struct {
	Username        string `form:"username"`
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"passwordConfirm"`
}

func GetSignupHandler(c echo.Context) error {
	return lib.Render(c, 200, view.Signup())
}

func PostSignupHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not sign up")
		}

		users, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		record := models.NewRecord(users)

		form := forms.NewRecordUpsert(app, record)

		form.LoadData(map[string]any{
			"username":        u.Username,
			"name":            u.Username,
			"email":           u.Email,
			"password":        u.Password,
			"passwordConfirm": u.ConfirmPassword,
		})
		// if err = form.LoadRequest(c.Request(), ""); err != nil {
		// 	log.Print(err)
		// 	return echo.NewHTTPError(http.StatusBadRequest, "Could not sign up")
		// }
		if err = form.Submit(); err != nil {
			log.Print(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Could not sign up")
		}

		return c.Redirect(http.StatusFound, "/login")
	}
}
