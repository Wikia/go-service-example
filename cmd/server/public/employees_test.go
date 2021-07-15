package public_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"

	"github.com/Wikia/go-example-service/internal/validator"

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
	req := httptest.NewRequest(http.MethodGet, "/example/employee/all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.GetAllEmployees(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
		assert.JSONEq(t, `[{"ID":1,"Name":"John Wick","City":"Atlanta"},{"ID":2,"Name":"Wade Winston Wilson","City":"New York"}]`, rec.Body.String())
	}
}

func TestGetAllEmployeesFail(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetAllEmployeesReturns(nil, errors.New("some error"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employee/all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.GetAllEmployees(c)
	if assert.Error(t, err) {
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
		assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
	}
}

func TestDeleteEmployee(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/example/employee/1", nil)
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

func TestDeleteEmployeeMissing(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)
	mockRepo.DeleteEmployeeReturns(gorm.ErrRecordNotFound)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/example/employee/5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.DeleteEmployee(c, 5)
	if assert.Error(t, err) {
		assert.Equal(t, http.StatusNotFound, err.(*echo.HTTPError).Code)
		assert.Equal(t, 1, mockRepo.DeleteEmployeeCallCount())
		_, id := mockRepo.DeleteEmployeeArgsForCall(0)
		assert.EqualValues(t, 5, id)
	}
}

func TestFindEmployeeByID(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetEmployeeReturns(&stubEmployees[0], nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employee/1", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/example/employee/2", nil)
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

func TestCreateEmployee(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	e := echo.New()
	e.Validator = &validator.EchoValidator{}
	payload, err := json.Marshal(stubEmployees[0])
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/example/employee/", bytes.NewBuffer(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.CreateEmployee(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, 1, mockRepo.AddEmployeeCallCount())
		_, ret := mockRepo.AddEmployeeArgsForCall(0)
		assert.EqualValues(t, &stubEmployees[0], ret)
		assert.Empty(t, rec.Body.String())
	}
}
