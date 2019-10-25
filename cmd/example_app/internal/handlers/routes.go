package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *zap.SugaredLogger) http.Handler {

	r := chi.NewRouter()

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", http.HandlerFunc(Hello))
	})

	return r
}
