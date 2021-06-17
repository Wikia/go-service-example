package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Readiness(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}
