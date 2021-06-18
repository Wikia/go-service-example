package main

import (
	"expvar"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Wikia/go-example-service/cmd/server/handlers"
	"github.com/Wikia/go-example-service/cmd/server/metrics"
	"github.com/Wikia/go-example-service/cmd/server/models"
	"github.com/Wikia/go-example-service/internal/tracing"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

const AppName = "example"

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
	} else {
		logCfg = zap.NewProductionConfig()
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

	db, err := gorm.Open(sqlite.Open(cfg.DB.Database), &gorm.Config{})
	if err != nil {
		sugared.With("error", err).Panic("failed to connect database")
	}

	//Init for this example
	var numEmployees int64
	db.Model(&models.Employee{}).Count(&numEmployees)
	if numEmployees == 0 {
		if err = models.InitData(db); err != nil {
			sugared.With("error", err).Error("error while initializing sample database data")
		}
	}

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	sugared.With("version", build).Info("Started : Application initializing")
	defer sugared.Info("Application terminated")

	// metrics
	registry := prometheus.DefaultRegisterer
	metrics.RegisterMetrics(prometheus.WrapRegistererWithPrefix(fmt.Sprintf("%s_", AppName), registry))

	// tracer
	tracer, closer, err := tracing.InitJaegerTracer(AppName, sugared, registry)
	if err != nil {
		return errors.Wrap(err, "error initializing tracer")
	}
	defer func() {
		err := closer.Close()
		if err != nil {
			sugared.With("error", err).Error("could not close tracer")
		}
	}()

	err = db.Use(gormopentracing.New(gormopentracing.WithTracer(tracer)))
	if err != nil {
		sugared.With("error", err).Error("could not initialize tracing for the database")
	}

	go func() {
		internal := handlers.Internal(logger)
		internal.HideBanner = true // no need to see it twice
		internal.HidePort = cfg.Environment != "dev"
		err = internal.Start(cfg.Web.InternalHost)
		if err != nil {
			sugared.With("error", err).Fatal("error starting internal server")
		}
	}()

	api := handlers.API(logger, tracer, AppName, db)
	api.HideBanner = cfg.Environment != "dev"
	api.HidePort = cfg.Environment != "dev"

	err = api.Start(cfg.Web.APIHost)
	if err != nil {
		sugared.With("error", err).Fatal("error starting server")
	}

	return nil
}
