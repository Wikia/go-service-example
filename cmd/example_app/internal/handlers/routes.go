package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	logmiddleware "github.com/harnash/go-middlewares/logger"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, logger *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()
	r.Use(
		logmiddleware.InContext(
			logmiddleware.WithLogger(func() (*zap.SugaredLogger, error){ return logger, nil })),
	)

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", Hello)
	})

	return r
}
