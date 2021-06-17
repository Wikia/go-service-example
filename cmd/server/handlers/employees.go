package handlers

import (
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
)

type Employee struct {
	Id   int
	Name string
	City string
}

func All(db *gorm.DB) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := zap.S()
		logger.Info("Fetching list of all employees")

		people, err := models.AllEmployees(db)

		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, people)
	}
}
