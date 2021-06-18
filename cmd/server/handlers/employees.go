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
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, people)
	}
}

func CreateEmployee(db *gorm.DB) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := zap.S()
		e := &models.Employee{}
		if err := ctx.Bind(e); err != nil {
			return err
		}
		logger.With("employee", e).Info("creating new employee")
		if err := models.AddEmployee(db, e); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return ctx.NoContent(http.StatusAccepted)
	}
}

func GetEmployee(db *gorm.DB) func (ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := zap.S()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("looking up employee")
		e, err := models.GetEmployee(db, employeeId)
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "object with given id not found")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, e)
	}
}

func DeleteEmployee(db *gorm.DB) func (ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := zap.S()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("deleting employee")
		err := models.DeleteEmployee(db, employeeId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.NoContent(http.StatusAccepted)
	}
}