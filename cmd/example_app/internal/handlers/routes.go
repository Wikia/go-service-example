package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/harnash/go-middlewares/logging"
	"github.com/harnash/go-middlewares/http_metrics"
	"github.com/harnash/go-middlewares/recovery"
	"github.com/harnash/go-middlewares/tracing"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, logger *zap.SugaredLogger, tracer opentracing.Tracer, db *gorm.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(
		logging.InContext(logging.WithLogger(func() (*zap.SugaredLogger, error) { return logger, nil })),
		recovery.PanicCatch(),
		logging.AccessLog(),
		tracing.Traced(tracing.WithTracer(tracer)),
		logging.InContext(
			logging.WithLogger(func() (*zap.SugaredLogger, error) { return logger, nil })),
	)

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", http_metrics.Measured(http_metrics.WithName("hello"))(http.HandlerFunc(Hello)).ServeHTTP)
		r.Route("/employee", func(r chi.Router) {
			r.Get("/all", http_metrics.Measured(
				http_metrics.WithName("all_employee"))(http.HandlerFunc(All(db))).ServeHTTP)
		})
	})

	return r
}
