package admin

import (
	"net/http"

	"github.com/pkg/errors"

	internalHandlers "github.com/Wikia/go-example-service/internal/handlers"
	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// NewInternalServer constructs an echo server with all application routes defined.
func NewInternalServer(logger *zap.Logger, swagger *openapi3.T) *echo.Echo {
	r := echo.New()

	r.Use(
		middleware.RemoveTrailingSlash(),
		logging.EchoLogger(logger),
		middleware.RecoverWithConfig(middleware.RecoverConfig{LogLevel: log.ERROR}),
	)

	health := r.Group("/health")
	{
		health.GET("/alive", internalHandlers.HealthCheck)
		health.GET("/ready", internalHandlers.Readiness)
	}

	r.GET("/metrics", func(ctx echo.Context) error {
		promhttp.Handler().ServeHTTP(ctx.Response(), ctx.Request())

		return nil
	})

	r.GET("/swagger", func(ctx echo.Context) error {
		data, err := swagger.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "error marshaling swagger spec")
		}

		return ctx.JSONBlob(http.StatusOK, data)
	})

	return r
}
