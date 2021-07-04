package public

import (
	"github.com/Wikia/go-example-service/cmd/models/employee"
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
	"gorm.io/gorm"
)

type APIServer struct {
	employeeRepo employee.Repository
}

// NewAPIServer constructs a public echo server with all application routes defined.
func NewAPIServer(logger *zap.Logger, tracer opentracing.Tracer, appName string, db *gorm.DB, swagger *openapi3.T) *echo.Echo {
	wrapper := APIServer{employeeRepo: employee.NewSQLRepository(db)}
	r := echo.New()
	traceConfig := jaegertracing.DefaultTraceConfig
	traceConfig.ComponentName = appName
	traceConfig.Tracer = tracer

	traceMiddleware := jaegertracing.TraceWithConfig(traceConfig)
	promMetrics := prometheus.NewPrometheus("http", func(c echo.Context) bool { return false })

	r.Use(
		traceMiddleware,
		logging.EchoLogger(logger),
		openapimiddleware.OapiRequestValidator(swagger),
		middleware.RecoverWithConfig(middleware.RecoverConfig{LogLevel: log.ERROR}),
	)

	promMetrics.Use(r)
	// request/form validation
	r.Validator = &validator.EchoValidator{}

	openapi.RegisterHandlers(r, &wrapper)

	return r
}
