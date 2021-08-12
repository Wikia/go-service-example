package public

import (
	"github.com/Wikia/go-example-service/cmd/models"
	"github.com/Wikia/go-example-service/cmd/openapi"
	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/Wikia/go-example-service/internal/validator"
	openapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type APIServer struct {
	employeeRepo models.Repository
}

func NewAPIServer(repository models.Repository) *APIServer {
	return &APIServer{repository}
}

// NewPublicAPI constructs a public echo server with all application routes defined.
func NewPublicAPI(logger *zap.Logger, tracer opentracing.Tracer, appName string, repository models.Repository, swagger *openapi3.T) *echo.Echo {
	wrapper := NewAPIServer(repository)
	r := echo.New()
	traceConfig := jaegertracing.DefaultTraceConfig
	traceConfig.ComponentName = appName
	traceConfig.Tracer = tracer

	traceMiddleware := jaegertracing.TraceWithConfig(traceConfig)
	promMetrics := prometheus.NewPrometheus("http", func(c echo.Context) bool { return false })

	r.Use(
		middleware.RemoveTrailingSlash(),
		traceMiddleware,
		logging.EchoLogger(logger),
	)

	if swagger != nil {
		r.Use(openapimiddleware.OapiRequestValidator(swagger))
	}
	r.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{LogLevel: log.ERROR}))

	promMetrics.Use(r)
	// request/form validation
	r.Validator = &validator.EchoValidator{}

	openapi.RegisterHandlers(r, wrapper)

	return r
}
