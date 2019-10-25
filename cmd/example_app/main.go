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
	"github.com/Wikia/go-example-service/internal/tracing"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

const APP_NAME = "example"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg struct {
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
		}
	}

	if err := conf.Parse(os.Args[1:], APP_NAME, &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage(APP_NAME, &cfg)
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
	var err error

	if cfg.Logging.Type != "dev" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return errors.Wrap(err, "could not initialize logger")
	}
	sugared := logger.Sugar().With("appname", "example")

	sugared.With("config", cfg).Info("Starting service")

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

	// tracer
	tracer, closer, err := tracing.InitJaegerTracer(APP_NAME, sugared, registry)
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
		Handler:      handlers.API(shutdown, sugared, tracer),
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
