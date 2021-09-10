package public

import (
	"errors"
	"net/http"

	"github.com/Wikia/go-commons/logging"
	"github.com/Wikia/go-example-service/api"
	"github.com/Wikia/go-example-service/internal/database"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func employeeDBModelFromCreateRequest(e api.CreateEmployeeRequest) *database.EmployeeDBModel {
	return &database.EmployeeDBModel{
		Name: e.Name,
		City: e.City,
	}
}

func employeeResponseFromDBModel(e database.EmployeeDBModel) *api.EmployeeResponse {
	return &api.EmployeeResponse{
		ID:   e.ID,
		Name: e.Name,
		City: e.City,
	}
}

func (s APIServer) GetAllEmployees(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx)
	logger.Info("Fetching list of all employees")

	people, err := s.employeeRepo.GetAllEmployees(ctx.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	response := make([]*api.EmployeeResponse, len(people))
	for pos, e := range people {
		response[pos] = employeeResponseFromDBModel(e)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (s APIServer) CreateEmployee(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	e := api.CreateEmployeeRequest{}

	if err := ctx.Bind(&e); err != nil {
		return err
	}

	if err := ctx.Validate(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	logger.With("employee", e).Info("creating new employee")

	dbEmployee := employeeDBModelFromCreateRequest(e)

	if err := s.employeeRepo.AddEmployee(ctx.Request().Context(), dbEmployee); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s APIServer) FindEmployeeByID(ctx echo.Context, employeeID int64) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	logger.With("id", employeeID).Info("looking up employee")
	e, err := s.employeeRepo.GetEmployee(ctx.Request().Context(), employeeID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "object with given id not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, employeeResponseFromDBModel(*e))
}

func (s APIServer) DeleteEmployee(ctx echo.Context, employeeID int64) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	logger.With("id", employeeID).Info("deleting employee")
	err := s.employeeRepo.DeleteEmployee(ctx.Request().Context(), employeeID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "object with given id not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusAccepted)
}
