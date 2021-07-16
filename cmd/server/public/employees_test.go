package public_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/Wikia/go-example-service/cmd/models/employee"

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
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	mockRepo.GetAllEmployeesReturns(stubEmployees, nil)

	req := httptest.NewRequest(http.MethodGet, "/example/employee/all", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
	assert.JSONEq(t, `[{"ID":1,"Name":"John Wick","City":"Atlanta"},{"ID":2,"Name":"Wade Winston Wilson","City":"New York"}]`, rec.Body.String())
}

func TestGetAllEmployeesFail(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	mockRepo.GetAllEmployeesReturns(nil, errors.New("some error"))

	req := httptest.NewRequest(http.MethodGet, "/example/employee/all", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 1, mockRepo.GetAllEmployeesCallCount())
}

func TestDeleteEmployee(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	req := httptest.NewRequest(http.MethodDelete, "/example/employee/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, 1, mockRepo.DeleteEmployeeCallCount())
	_, id := mockRepo.DeleteEmployeeArgsForCall(0)
	assert.EqualValues(t, 1, id)
}

func TestDeleteEmployeeMissing(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	mockRepo.DeleteEmployeeReturns(gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/example/employee/5", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, 1, mockRepo.DeleteEmployeeCallCount())
	_, id := mockRepo.DeleteEmployeeArgsForCall(0)
	assert.EqualValues(t, 5, id)
}

func TestFindEmployeeByID(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	mockRepo.GetEmployeeReturns(&stubEmployees[0], nil)

	req := httptest.NewRequest(http.MethodGet, "/example/employee/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, mockRepo.GetEmployeeCallCount())
	_, id := mockRepo.GetEmployeeArgsForCall(0)
	assert.EqualValues(t, 1, id)
	assert.JSONEq(t, `{"ID":1,"Name":"John Wick","City":"Atlanta"}`, rec.Body.String())
}

func TestFindEmployeeByIDMissing(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	mockRepo.GetEmployeeReturns(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/example/employee/2", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, 1, mockRepo.GetEmployeeCallCount())
	_, id := mockRepo.GetEmployeeArgsForCall(0)
	assert.EqualValues(t, 2, id)
	assert.JSONEq(t, `{"message":"object with given id not found"}`, rec.Body.String())
}

func TestCreateEmployee(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	payload, err := json.Marshal(stubEmployees[0])
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/example/employee", bytes.NewBuffer(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, mockRepo.AddEmployeeCallCount())
	_, ret := mockRepo.AddEmployeeArgsForCall(0)
	assert.EqualValues(t, &stubEmployees[0], ret)
	assert.Empty(t, rec.Body.String())
}

func TestCreateEmployeeInvalid(t *testing.T) {
	mockRepo := &employeefakes.FakeRepository{}
	e := public.NewPublicAPI(zap.L(), opentracing.NoopTracer{}, "test-server", mockRepo, nil)

	badEmployee := employee.Employee{Name: "Joker"}
	payload, err := json.Marshal(badEmployee)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/example/employee", bytes.NewBuffer(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Regexp(t, `Error:Field validation`, rec.Body.String())
	assert.Equal(t, 0, mockRepo.AddEmployeeCallCount())
}
