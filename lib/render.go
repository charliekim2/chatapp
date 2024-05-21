package lib

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
)

func Render(c echo.Context, status int, t templ.Component) error {
	c.Response().Writer.WriteHeader(status)

	err := t.Render(context.Background(), c.Response().Writer)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Error rendering template")
	}

	return nil
}
