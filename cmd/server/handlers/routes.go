package handlers

import (
	"github.com/brpaz/echozap"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// API constructs an http.Handler with all application routes defined.
func API(logger *zap.Logger, tracer opentracing.Tracer, appName string, db *gorm.DB) *echo.Echo {
	r := echo.New()
	traceConfig := jaegertracing.DefaultTraceConfig
	traceConfig.ComponentName = appName
	traceConfig.Tracer = tracer

	traceMiddleware := jaegertracing.TraceWithConfig(traceConfig)

	r.Use(
		echozap.ZapLogger(logger),
		middleware.RecoverWithConfig(middleware.RecoverConfig{LogLevel: log.ERROR}),
		traceMiddleware,
	)

	example := r.Group("/example")
	{
		example.GET("/hello", Hello)
		employee := example.Group("/employee")
		{
			employee.GET("/all", All(db))
			employee.PUT("/", CreateEmployee(db))
			employee.GET("/:id", GetEmployee(db))
			employee.DELETE("/:id", DeleteEmployee(db))
		}
	}

	return r
}
