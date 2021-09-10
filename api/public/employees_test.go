package public_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wikia/go-example-service/internal/database"
	"github.com/Wikia/go-example-service/internal/database/databasefakes"
	"github.com/pkg/errors"

	"github.com/Wikia/go-example-service/internal/validator"

	"gorm.io/gorm"

	"github.com/Wikia/go-example-service/internal/logging"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"

	"github.com/stretchr/testify/assert"

	"github.com/Wikia/go-example-service/api/public"
)

var stubEmployees = []database.EmployeeDBModel{
	{
		ID: 0, Name: "John Wick", City: "Atlanta",
	},
	{
		ID: 1, Name: "Wade Winston Wilson", City: "New York",
	},
}

func TestGetAllEmployees(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
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
		assert.JSONEq(t, `[{"id":0,"name":"John Wick","city":"Atlanta"},{"id":1,"name":"Wade Winston Wilson","city":"New York"}]`, rec.Body.String())
	}
}

func TestGetAllEmployeesFail(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetAllEmployeesReturns(nil, errors.New("some error"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employee/all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.GetAllEmployees(c)
	var httpError *echo.HTTPError
	if assert.ErrorAs(t, err, &httpError) {
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
	}
}

func TestDeleteEmployee(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
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
	mockRepo := &databasefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)
	mockRepo.DeleteEmployeeReturns(gorm.ErrRecordNotFound)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/example/employee/5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.DeleteEmployee(c, 5)
	var httpError *echo.HTTPError
	if assert.ErrorAs(t, err, &httpError) {
		assert.Equal(t, http.StatusNotFound, httpError.Code)
		assert.Equal(t, 1, mockRepo.DeleteEmployeeCallCount())
		_, id := mockRepo.DeleteEmployeeArgsForCall(0)
		assert.EqualValues(t, 5, id)
	}
}

func TestFindEmployeeByID(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
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
		assert.JSONEq(t, `{"id":0,"name":"John Wick","city":"Atlanta"}`, rec.Body.String())
	}
}

func TestFindEmployeeByIDMissing(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	mockRepo.GetEmployeeReturns(nil, gorm.ErrRecordNotFound)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/example/employee/2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err := server.FindEmployeeByID(c, 2)
	var httpError *echo.HTTPError
	if assert.ErrorAs(t, err, &httpError) {
		assert.Equal(t, http.StatusNotFound, httpError.Code)
		assert.Equal(t, 1, mockRepo.GetEmployeeCallCount())
		_, id := mockRepo.GetEmployeeArgsForCall(0)
		assert.EqualValues(t, 2, id)
		assert.Empty(t, rec.Body.String())
	}
}

func TestCreateEmployee(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	e := echo.New()
	e.Validator = &validator.EchoValidator{}
	payload, err := json.Marshal(stubEmployees[0])
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/example/employee", bytes.NewBuffer(payload))
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

func TestCreateEmployeeInvalid(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)
	badEmployee := database.EmployeeDBModel{Name: "Joker"}

	e := echo.New()
	e.Validator = &validator.EchoValidator{}
	payload, err := json.Marshal(badEmployee)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/example/employee", bytes.NewBuffer(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	err = server.CreateEmployee(c)
	var httpError *echo.HTTPError
	if assert.ErrorAs(t, err, &httpError) {
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Regexp(t, `Error:Field validation`, httpError.Message)
		assert.Equal(t, 0, mockRepo.AddEmployeeCallCount())
	}
}
