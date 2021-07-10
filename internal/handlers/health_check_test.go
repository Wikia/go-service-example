package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wikia/go-example-service/internal/handlers"

	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"
)

func TestHealthCheckHandler(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()

	if c := e.NewContext(req, rec); assert.NoError(t, handlers.HealthCheck(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}
