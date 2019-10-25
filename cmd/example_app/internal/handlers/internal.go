package handlers

import (
	"net/http"
	"os"

	ihandlers "github.com/Wikia/go-example-service/internal/handlers"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func Internal(shutdown chan os.Signal, log *zap.SugaredLogger) http.Handler {

	r := chi.NewRouter()
	r.Get("/healhcheck", http.HandlerFunc(ihandlers.HealthCheck))
	r.Get("/readiness", http.HandlerFunc(ihandlers.Readiness))
	return r
}
