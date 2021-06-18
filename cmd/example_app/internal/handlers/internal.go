package handlers

import (
	"net/http"
	"os"

	internalHandlers "github.com/Wikia/go-example-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func Internal(shutdown chan os.Signal, log *zap.SugaredLogger) http.Handler {

	r := chi.NewRouter()
	r.Get("/healthcheck", http.HandlerFunc(internalHandlers.HealthCheck))
	r.Get("/readiness", http.HandlerFunc(internalHandlers.Readiness))
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	return r
}
