package handlers

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

// API constructs an http.Handler with all application routes defined.
func API(logger *zap.Logger, appName string, db *gorm.DB) *gin.Engine {
	m := ginmetrics.GetMonitor()
	r := gin.New()
	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
		otelgin.Middleware(appName),
	)
	m.Use(r)

	example := r.Group("/example")
	{
		example.GET("/hello", Hello)
		employee := example.Group("/employee")
		{
			employee.GET("/all", AllEmployees(db))
			employee.PUT("/", CreateEmployee(db))
		}
	}

	return r
}
