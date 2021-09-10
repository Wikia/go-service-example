package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	dbcommons "github.com/Wikia/go-commons/database"
	"github.com/Wikia/go-commons/tracing"
	"github.com/Wikia/go-example-service/internal/database"
	"github.com/labstack/echo/v4"
	dblogger "gorm.io/gorm/logger"

	"github.com/Wikia/go-example-service/api/admin"
	"github.com/Wikia/go-example-service/api/public"
	"github.com/Wikia/go-example-service/cmd/openapi"

	"github.com/Wikia/go-example-service/metrics"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormopentracing "gorm.io/plugin/opentracing"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

// AppName should hold unique name of your service.
// Please be aware that this is also used as a prefix for environment variables used in config
const AppName = "example"
const ShutdownTimeout = 10

func main() {
	if err := run(); err != nil {
		zap.L().With(zap.Error(err)).Error("error running service")
		os.Exit(1)
	}
}

func startServer(logger *zap.Logger, e *echo.Echo, host string) {
	if err := e.Start(host); err != nil && errors.Is(err, http.ErrServerClosed) {
		logger.With(zap.Error(err)).With(zap.String("host", host)).Fatal("error starting/running server")
	}
}

func run() error {
	var cfg struct {
		Environment string `conf:"default:prod,help:name of the environment app is running in (prod/dev/localhost)"`
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
			Type  string `conf:"default:prod,help:can be one of prod/dev/localhost"`
			Level string `conf:"default:info"`
		}
		DB struct {
			Driver          string `conf:"default:sqlite3"`
			User            string `conf:"default:root"`
			Password        string `conf:"default:root"`
			Host            string `conf:"default:localhost"`
			Sources         []string
			Replicas        []string
			ConnMaxIdleTime time.Duration `conf:"default:1h"`
			ConnMaxLifeTime time.Duration `conf:"default:12h"`
			MaxIdleConns    int           `conf:"default:10"` // tune this to your needs
			MaxOpenConns    int           `conf:"default:20"` // this as well
		}
	}

	if err := conf.Parse(os.Args[1:], AppName, &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
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

	var (
		logger *zap.Logger
		logCfg zap.Config
		err    error
	)

	const LocalhostEnv = "localhost"
	if cfg.Logging.Type == "dev" || cfg.Logging.Type == LocalhostEnv {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	if cfg.Environment == LocalhostEnv {
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

	logger = logger.With(
		zap.String("appname", AppName),
		zap.String("version", version),
		zap.String("git_commit", commit),
		zap.String("build_date", date),
		zap.String("build_by", builtBy),
		zap.String("environment", cfg.Environment),
		zap.String("datacenter", cfg.Datacenter),
		zap.String("pod_name", cfg.K8S.PodName),
	)

	logger.With(zap.Reflect("config", cfg)).Info("starting service")

	zap.ReplaceGlobals(logger)

	// =========================================================================
	// DB
	db, err := dbcommons.GetConnection(logger, dblogger.Info, cfg.DB.Sources, cfg.DB.Replicas, cfg.DB.ConnMaxIdleTime, cfg.DB.ConnMaxLifeTime, cfg.DB.MaxIdleConns, cfg.DB.MaxOpenConns)

	if err != nil {
		logger.With(zap.Error(err)).Panic("could not connect to the database")
	}

	// Init for this example
	res := db.Raw("SHOW TABLES LIKE 'employees'")

	var result string
	err = res.Row().Scan(&result)

	if err != nil || len(result) == 0 {
		logger.Info("no tables found - initializing database")

		if err = database.InitData(db); err != nil {
			logger.With(zap.Error(err)).Warn("could not initialize database")
		}
	}

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(commit)
	expvar.NewString("version").Set(version)
	logger.Info("started: Application initializing")

	defer logger.Info("application terminated")

	// metrics
	registry := prometheus.DefaultRegisterer
	metrics.RegisterMetrics(prometheus.WrapRegistererWithPrefix(fmt.Sprintf("%s_", AppName), registry))

	// tracer
	tracer, closer, err := tracing.InitJaegerTracer(AppName, logger.Sugar(), registry)
	if err != nil {
		return errors.Wrap(err, "error initializing tracer")
	}

	defer func() {
		err := closer.Close()
		if err != nil {
			logger.With(zap.Error(err)).Error("could not close tracer")
		}
	}()

	err = db.Use(gormopentracing.New(gormopentracing.WithTracer(tracer)))

	if err != nil {
		logger.With(zap.Error(err)).Error("could not initialize tracing for the database")
	}

	// swagger
	swagger, err := openapi.GetSwagger()

	if err != nil {
		logger.With(zap.Error(err)).Fatal("could not load Swagger Spec")
	}

	swagger.Servers = nil

	internalAPI := admin.NewInternalAPI(logger, swagger)
	internalAPI.HideBanner = cfg.Environment != LocalhostEnv
	internalAPI.HidePort = cfg.Environment != LocalhostEnv
	go startServer(logger, internalAPI, cfg.Web.InternalHost)

	sqlRepo := database.NewSQLRepository(db)
	publicAPI := public.NewPublicAPI(logger, tracer, AppName, sqlRepo, swagger)
	publicAPI.HideBanner = true // no need to see it twice
	publicAPI.HidePort = cfg.Environment != LocalhostEnv
	go startServer(logger, publicAPI, cfg.Web.APIHost)

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout*time.Second)
	defer cancel()
	if err := publicAPI.Shutdown(ctx); err != nil {
		publicAPI.Logger.Fatal(err)
	}
	if err := internalAPI.Shutdown(ctx); err != nil {
		internalAPI.Logger.Fatal(err)
	}

	return nil
}
