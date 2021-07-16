package public_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wikia/go-example-service/cmd/models/employee/employeefakes"
	"github.com/Wikia/go-example-service/cmd/server/public"
	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/Wikia/go-example-service/internal/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGreet(t *testing.T) {
	t.Parallel()
	mockRepo := &employeefakes.FakeRepository{}
	server := public.NewAPIServer(mockRepo)

	e := echo.New()
	e.Validator = &validator.EchoValidator{}

	req := httptest.NewRequest(http.MethodPut, "/example/greet", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logging.AddToContext(c, zap.L())

	if assert.NoError(t, server.Greet(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"Text": "Hello World!"}`, rec.Body.String())
	}
}
