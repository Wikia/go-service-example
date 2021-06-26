package handlers

import (
	"github.com/Wikia/go-example-service/internal/logging"
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func All(db *gorm.DB) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := logging.FromEchoContext(ctx)
		logger.Info("Fetching list of all employees")

		people, err := models.AllEmployees(ctx.Request().Context(), db)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, people)
	}
}

func CreateEmployee(db *gorm.DB) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := logging.FromEchoContext(ctx).Sugar()
		e := &models.Employee{}
		if err := ctx.Bind(e); err != nil {
			return err
		}
		if err := ctx.Validate(e); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		logger.With("employee", e).Info("creating new employee")
		if err := models.AddEmployee(ctx.Request().Context(), db, e); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return ctx.NoContent(http.StatusAccepted)
	}
}

func GetEmployee(db *gorm.DB) func (ctx echo.Context) error {
	return func(ctx echo.Context) error {
		logger := logging.FromEchoContext(ctx).Sugar()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("looking up employee")
		e, err := models.GetEmployee(ctx.Request().Context(), db, employeeId)
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
		logger := logging.FromEchoContext(ctx).Sugar()
		employeeId := ctx.Param("id")
		logger.With("id", employeeId).Info("deleting employee")
		err := models.DeleteEmployee(ctx.Request().Context(), db, employeeId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.NoContent(http.StatusAccepted)
	}
}