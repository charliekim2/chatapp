package handler

import (
	"log"
	"net/http"

	"github.com/charliekim2/chatapp/db"
	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/model"
	"github.com/charliekim2/chatapp/view/layout"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
)

func EditProfileHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		user, err := db.ReadUser(app, authRecord.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Error getting user info")
		}

		return lib.Render(c, 200, layout.EditProfile(user))
	}
}

func UpdateUserHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		user := new(model.DBUser)
		err := c.Bind(user)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Could not edit profile")
		}
		user.Id = authRecord.Id

		err = db.UpdateUser(app, user)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Could not edit profile")
		}

		users, err := app.Dao().FindCollectionByNameOrId("users")

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Error querying users")
		}

		form := forms.NewRecordPasswordLogin(app, users)
		form.Identity = user.Name
		form.Password = user.Password
		authRecord, err = form.Submit()

		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Error validating login")
		}

		token, err := tokens.NewRecordAuthToken(app, authRecord)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Error generating login token")
		}

		c.SetCookie(&http.Cookie{
			Name:     "pb_auth",
			Value:    token,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			HttpOnly: true,
			MaxAge:   int(app.Settings().RecordAuthToken.Duration),
			Path:     "/",
		})

		return c.Redirect(http.StatusFound, "/editprofile")
	}
}

func UploadAvatarHandler(app *pocketbase.PocketBase) func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/editprofile")
	}
}
