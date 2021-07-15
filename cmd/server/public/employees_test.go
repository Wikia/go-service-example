package public_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		1, "John Wick", "Atlanta",
	},
	{
		2, "Wade Winston Wilson", "New York",
	},
}

func TestGetAllEmployees(t *testing.T) {
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
		assert.JSONEq(t, `[{"ID":1,"Name":"John Wick","City":"Atlanta"},{"ID":2,"Name":"Wade Winston Wilson","City":"New York"}]`, rec.Body.String())
	}
}
