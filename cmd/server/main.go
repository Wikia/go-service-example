package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Wikia/go-example-service/cmd/server/handlers"
	"github.com/Wikia/go-example-service/cmd/server/metrics"
	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/Wikia/go-example-service/internal/tracing"
	"github.com/ardanlabs/conf"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

const AppName = "example"

var tracer = otel.Tracer(AppName)

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg struct {
		Environment string `conf:"default:prod,name of the environment app is running in (prod/dev/localhost)"`
		Datacenter  string `conf:"help:name of the environment app is running on"`
		K8S         struct {
			PodName string `conf:"help:name of the pod running the app"`
		}
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			InternalHost    string        `conf:"default:0.0.0.0:4000"`
			DebugHost       string        `conf:"default:0.0.0.0:5000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Logging struct {
			Type  string `conf:"default:prod"`
			Level string `conf:"default:info"`
		}
		DB struct {
			Driver   string `conf:"default:sqlite3"`
			User     string `conf:"default:root"`
			Password string `conf:"default:root"`
			Host     string `conf:"default:localhost"`
			Database string `conf:"default:test.db"`
		}
	}

	if err := conf.Parse(os.Args[1:], AppName, &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage(AppName, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// Logging

	var logger *zap.Logger
	var logCfg zap.Config
	var err error

	if cfg.Logging.Type == "dev" || cfg.Logging.Type == "localhost" {
		logCfg = zap.NewDevelopmentConfig()
		gin.SetMode(gin.DebugMode)
	} else {
		logCfg = zap.NewProductionConfig()
		gin.SetMode(gin.ReleaseMode)
	}

	if cfg.Environment == "localhost" {
		logCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}

	logLevel := zap.InfoLevel
	err = logLevel.Set(cfg.Logging.Level)
	if err == nil {
		logCfg.Level = zap.NewAtomicLevelAt(logLevel)
		logger, err = logCfg.Build()
	}

	if err != nil {
		panic(fmt.Sprintf("could not initialize log: %v", err))
	}
	sugared := logger.Sugar().With("appname", AppName, "environment", cfg.Environment, "datacenter", cfg.Datacenter, "pod_name", cfg.K8S.PodName)

	sugared.With("config", cfg).Info("Starting service")

	zap.ReplaceGlobals(logger)

	// =========================================================================
	// DB

	db, err := gorm.Open(cfg.DB.Driver, cfg.DB.Database)
	if err != nil {
		sugared.With("error", err).Panic("failed to connect database")
	}
	db.SetLogger(&logging.TracingLogger{Logger: sugared})
	if cfg.Environment == "dev" {
		db.LogMode(true)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			sugared.With("error", err).Error("error while closing database handler")
		}
	}()

	//Init for this example
	if !db.HasTable(models.Employee{}) {
		models.InitData(db)
	}

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	sugared.With("version", build).Info("Started : Application initializing")
	defer sugared.Info("Application terminated")

	// metrics
	exporter, ctrl := metrics.RegisterMetrics(AppName, cfg.Environment, sugared)
	defer func(c *controller.Controller, ctx context.Context) {
		err := c.Stop(ctx)
		if err != nil {
			sugared.With("error", err).Error("could not stop metrics controller")
		}
	}(ctrl, context.Background())

	// tracer
	tp := tracing.InitJaegerTracer(AppName, cfg.Environment, sugared)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			sugared.With("error", err).Error("error shutting down tracer provider")
		}
	}()

	go func() {
		internal := handlers.Internal(logger, exporter)
		err = internal.Run(cfg.Web.InternalHost)
		if err != nil {
			sugared.With("error", err).Fatal("error starting internal server")
		}
	}()

	api := handlers.API(logger, AppName, db)
	err = api.Run(cfg.Web.APIHost)
	if err != nil {
		sugared.With("error", err).Fatal("error starting server")
	}

	return nil
}