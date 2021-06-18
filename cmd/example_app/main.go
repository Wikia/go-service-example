package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wikia/go-example-service/cmd/example_app/internal/handlers"
	"github.com/Wikia/go-example-service/cmd/example_app/internal/metrics"
	"github.com/Wikia/go-example-service/cmd/example_app/internal/models"
	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/Wikia/go-example-service/internal/tracing"
	"github.com/ardanlabs/conf"
	"github.com/harnash/go-middlewares/http_metrics"
	"github.com/harnash/go-middlewares/recovery"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		Datacenter string `conf:"help:name of the environment app is running on"`
		K8S struct {
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
			Type string `conf:"default:prod"`
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

	// =========================================================================
	// DB

	db, err := gorm.Open(cfg.DB.Driver, cfg.DB.Database)
	if err != nil {
		sugared.With("error", err).Panic("failed to connect database")
	}
	db.SetLogger(&logging.TracingLogger{Logger: sugared})

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

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	sugared.Info("Started : Initializing debugging support")

	go func() {
		sugared.With("debug_host", cfg.Web.DebugHost).Info("Debug Listening")
		sugared.Infof("Debug Listener closed : %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Start API Service

	sugared.Info("Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// metrics
	registry := prometheus.DefaultRegisterer
	metrics.RegisterMetrics(prometheus.WrapRegistererWithPrefix(fmt.Sprintf("%s_", AppName), registry))
	err = http_metrics.RegisterDefaultMetrics(registry)
	if err != nil {
		sugared.With("error", err).Error("could not initialize http middleware metrics")
	}

	err = recovery.RegisterDefaultMetrics(registry)
	if err != nil {
		sugared.With("error", err).Error("could not initialize panics middleware metrics")
	}

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

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(shutdown, sugared, tracer, db),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		sugared.With("address", api.Addr).Info("API listening")
		serverErrors <- api.ListenAndServe()
		sugared.Infof("API closed")
	}()

	// =========================================================================
	// Start HealthCheck Service

	sugared.Info("Started : Initializing internal support")

	internal := http.Server{
		Addr:         cfg.Web.InternalHost,
		Handler:      handlers.Internal(shutdown, sugared),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests.
	go func() {
		sugared.With("address", internal.Addr).Info("Internal listening")
		sugared.Infof("Internal closed: %v", internal.ListenAndServe())
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		sugared.With("signal", sig).Info("Start shutdown")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			sugared.With("timeout", cfg.Web.ShutdownTimeout, "error", err).Warn("Graceful shutdown did not complete")
			err = api.Close()
		}
		sugared.With("signal", sig).Info("API shutdown")

		// Asking listener to shutdown and load shed.
		err = internal.Shutdown(ctx)
		if err != nil {
			sugared.With("timeout", cfg.Web.ShutdownTimeout, "error", err).Warn("Graceful shutdown did not complete")
			err = internal.Close()
		}
		sugared.With("signal", sig).Info("Internal shutdown")

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
