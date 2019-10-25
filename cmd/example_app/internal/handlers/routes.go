package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	logmiddleware "github.com/harnash/go-middlewares/logger"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, logger *zap.SugaredLogger) http.Handler {

	r := chi.NewRouter()
	r.With(
		logmiddleware.InContext(
			logmiddleware.WithLogger(func() (*zap.SugaredLogger, error){ return logger, nil })),
	)

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", Hello)
	})

	return r
}
