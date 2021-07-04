package public

import (
	"net/http"

	"github.com/Wikia/go-example-service/internal/logging"

	"github.com/Wikia/go-example-service/cmd/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (s APIServer) GetAllEmployees(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx)
	logger.Info("Fetching list of all employees")

	people, err := models.AllEmployees(ctx.Request().Context(), s.DB)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, people)
}

func (s APIServer) CreateEmployee(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	e := &models.Employee{}
	if err := ctx.Bind(e); err != nil {
		return err
	}
	if err := ctx.Validate(e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	logger.With("employee", e).Info("creating new employee")
	if err := models.AddEmployee(ctx.Request().Context(), s.DB, e); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return ctx.NoContent(http.StatusAccepted)
}

func (s APIServer) FindEmployeeByID(ctx echo.Context, employeeID int64) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	logger.With("id", employeeID).Info("looking up employee")
	e, err := models.GetEmployee(ctx.Request().Context(), s.DB, employeeID)
	if err == gorm.ErrRecordNotFound {
		return echo.NewHTTPError(http.StatusNotFound, "object with given id not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, e)
}

func (s APIServer) DeleteEmployee(ctx echo.Context, employeeID int64) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	logger.With("id", employeeID).Info("deleting employee")
	err := models.DeleteEmployee(ctx.Request().Context(), s.DB, employeeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusAccepted)
}
