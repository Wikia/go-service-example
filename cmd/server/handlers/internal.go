package handlers

import (
	internalHandlers "github.com/Wikia/go-example-service/internal/handlers"
	"github.com/brpaz/echozap"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Internal constructs an http.Handler with all application routes defined.
func Internal(logger *zap.Logger) *echo.Echo {
	r := echo.New()

	r.Use(
		echozap.ZapLogger(logger),
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

	return r
}
