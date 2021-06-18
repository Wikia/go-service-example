package handlers

import (
	internalHandlers "github.com/Wikia/go-example-service/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.uber.org/zap"
)

// Internal constructs an http.Handler with all application routes defined.
func Internal(log *zap.Logger, exporter *prometheus.Exporter) *gin.Engine {
	r := gin.New()

	health := r.Group("/health")
	{
		health.GET("/alive", internalHandlers.HealthCheck)
		health.GET("/ready", internalHandlers.Readiness)
	}

	r.GET("/metrics", func(ctx *gin.Context) {
		exporter.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return r
}
