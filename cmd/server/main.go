package main

import (
	"expvar"
	"fmt"
	"github.com/Wikia/go-example-service/internal/database"
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
	gormopentracing "gorm.io/plugin/opentracing"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

//AppName should hold unique name of your service.
//Please be aware that this is also used as a prefix for environment variables used in config
const AppName = "example"

func main() {
	if err := run(); err != nil {
		zap.L().With(zap.Error(err)).Error("error running service")
		os.Exit(1)
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
			Driver   string `conf:"default:sqlite3"`
			User     string `conf:"default:root"`
			Password string `conf:"default:root"`
			Host     string `conf:"default:localhost"`
			Sources  []string
			Replicas []string
			ConnMaxIdleTime time.Duration `conf:"default:1h"`
			ConnMaxLifeTime time.Duration `conf:"default:12h"`
			MaxIdleConns int `conf:"default:10"` // tune this to your needs
			MaxOpenConns int `conf:"default:20"` // this as well
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
	db, err := database.GetConnection(logger, cfg.DB.Sources, cfg.DB.Replicas, cfg.DB.ConnMaxIdleTime, cfg.DB.ConnMaxLifeTime, cfg.DB.MaxIdleConns, cfg.DB.MaxOpenConns)
	if err != nil {
		logger.With(zap.Error(err)).Panic("could not connect to the database")
	}

	//Init for this example
	res := db.Raw("SHOW TABLES LIKE 'employees'")
	var result string
	err = res.Row().Scan(&result)
	if err != nil || len(result) == 0 {
		logger.Info("no tables found - initializing database")
		if err = models.InitData(db); err != nil {
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

	go func() {
		internal := handlers.Internal(logger)
		internal.HideBanner = cfg.Environment != "localhost"
		internal.HidePort = cfg.Environment != "localhost"
		err = internal.Start(cfg.Web.InternalHost)
		if err != nil {
			logger.With(zap.Error(err)).Fatal("error starting internal server")
		}
	}()

	api := handlers.API(logger, tracer, AppName, db)
	api.HideBanner = true // no need to see it twice
	api.HidePort = cfg.Environment != "localhost"

	err = api.Start(cfg.Web.APIHost)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("error starting server")
	}

	return nil
}
