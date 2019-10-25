package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, log *log.Logger) http.Handler {

	r := chi.NewRouter()

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", http.HandlerFunc(Hello))
	})

	return r
}
