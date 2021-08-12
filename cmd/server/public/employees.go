package public

import (
	"errors"
	"net/http"

	"github.com/Wikia/go-example-service/cmd/models"

	"github.com/Wikia/go-example-service/internal/logging"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func employeeDbModelFromCreateRequest(e models.CreateEmployeeRequest) *models.EmployeeDbModel {
	return &models.EmployeeDbModel{
		Name: e.Name,
		City: e.City,
	}
}

func employeeResponseFromDbModel(e models.EmployeeDbModel) *models.EmployeeResponse {
	return &models.EmployeeResponse{
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

	response := make([]*models.EmployeeResponse, len(people))
	for pos, e := range people {
		response[pos] = employeeResponseFromDbModel(e)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (s APIServer) CreateEmployee(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx).Sugar()
	e := models.CreateEmployeeRequest{}

	if err := ctx.Bind(&e); err != nil {
		return err
	}

	if err := ctx.Validate(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	logger.With("employee", e).Info("creating new employee")

	dbEmployee := employeeDbModelFromCreateRequest(e)

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

	return ctx.JSON(http.StatusOK, employeeResponseFromDbModel(*e))
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
