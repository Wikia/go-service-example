package public_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wikia/go-commons/logging"
	"github.com/Wikia/go-commons/validator"
	"github.com/Wikia/go-example-service/api/public"
	"github.com/Wikia/go-example-service/internal/database/databasefakes"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGreet(t *testing.T) {
	t.Parallel()
	mockRepo := &databasefakes.FakeRepository{}
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
