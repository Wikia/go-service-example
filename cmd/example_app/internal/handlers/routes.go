package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	logmiddleware "github.com/harnash/go-middlewares/logger"
	metricsmiddleware "github.com/harnash/go-middlewares/metrics"
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
		logmiddleware.InContext(logmiddleware.WithLogger(func() (*zap.SugaredLogger, error) { return logger, nil })),
		recovery.PanicCatch(),
		logmiddleware.AccessLog(),
		tracing.Traced(tracing.WithTracer(tracer)),
		logmiddleware.InContext(
			logmiddleware.WithLogger(func() (*zap.SugaredLogger, error) { return logger, nil })),
	)

	r.Route("/example", func(r chi.Router) {
		r.Get("/hello", metricsmiddleware.Measured(metricsmiddleware.WithName("hello"))(http.HandlerFunc(Hello)).ServeHTTP)
		r.Route("/employee", func(r chi.Router) {
			r.Get("/all", metricsmiddleware.Measured(
				metricsmiddleware.WithName("all_employee"))(http.HandlerFunc(All(db))).ServeHTTP)
		})
	})

	return r
}
