package public_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"

	"github.com/Wikia/go-example-service/cmd/models/employee"

	"github.com/Wikia/go-example-service/internal/logging"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"

	"github.com/stretchr/testify/assert"

	"github.com/Wikia/go-example-service/cmd/models/employee/employeefakes"
	"github.com/Wikia/go-example-service/cmd/server/public"
)

var stubEmployees = []employee.Employee{
	{
		ID: 1, Name: "John Wick", City: "Atlanta",
	},
	{
		ID: 2, Name: "Wade Winston Wilson", City: "New York",
	},
}

func TestGetAllEmployees(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetAllEmployeesReturns(stubEmployees, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employees/all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.GetAllEmployees(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
		assert.JSONEq(t, `[{"ID":1,"Name":"John Wick","City":"Atlanta"},{"ID":2,"Name":"Wade Winston Wilson","City":"New York"}]`, rec.Body.String())
	}
}

func TestDeleteEmployee(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/example/employees/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.DeleteEmployee(c, 1)) {
		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Equal(t, 1, mockRepo.DeleteEmployeeCallCount())
		_, id := mockRepo.DeleteEmployeeArgsForCall(0)
		assert.EqualValues(t, 1, id)
	}
}

func TestFindEmployeeByID(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetEmployeeReturns(&stubEmployees[0], nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employees/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.FindEmployeeByID(c, 1)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, mockRepo.GetEmployeeCallCount())
		_, id := mockRepo.GetEmployeeArgsForCall(0)
		assert.EqualValues(t, 1, id)
		assert.JSONEq(t, `{"ID":1,"Name":"John Wick","City":"Atlanta"}`, rec.Body.String())
	}
}

func TestFindEmployeeByIDMissing(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetEmployeeReturns(nil, gorm.ErrRecordNotFound)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employees/2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.FindEmployeeByID(c, 2)
	if assert.Error(t, err) {
		assert.Equal(t, http.StatusNotFound, err.(*echo.HTTPError).Code)
		assert.Equal(t, 1, mockRepo.GetEmployeeCallCount())
		_, id := mockRepo.GetEmployeeArgsForCall(0)
		assert.EqualValues(t, 2, id)
		assert.Empty(t, rec.Body.String())
	}
}
