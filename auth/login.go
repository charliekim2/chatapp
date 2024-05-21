package auth

import (
	"fmt"
	"net/http"

	"github.com/charliekim2/chatapp/lib"
	"github.com/charliekim2/chatapp/view"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/tokens"
)

func GetLoginHandler(c echo.Context) error {
	return lib.Render(c, 200, view.Login())
}

func PostLoginHandler(app *pocketbase.PocketBase) func(echo.Context) error {
	return func(c echo.Context) error {
		users, err := app.Dao().FindCollectionByNameOrId("users")

		if err != nil {
			// todo: custom error page
			return echo.NewHTTPError(http.StatusNotFound, "Error querying users")
		}

		form := forms.NewRecordPasswordLogin(app, users)
		c.Bind(form)
		authRecord, err := form.Submit()

		if err != nil {
			fmt.Println(err)
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

		return c.Redirect(http.StatusFound, "/")
	}
}
