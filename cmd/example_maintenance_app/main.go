package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Wikia/go-example-service/cmd/example_maintenance_app/internal"
	"github.com/ardanlabs/conf"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go.uber.org/zap/zapcore"
)

const AppName = "example_maintenance_app"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg struct {
		DryRun bool `conf:"default:true,help:do not make any changes - just display intended changes"`
		Delay time.Duration `conf:"default:5s,help:delay before starting applying changes"`
		Environment string `conf:"default:prod,name of the environment app is running in (prod/dev/localhost)"`
		Datacenter string `conf:"help:name of the environment app is running on"`
		K8S struct {
			PodName string `conf:"help:name of the pod running the app"`
		}

		Logging struct {
			Type string `conf:"default:prod,help:set logging format (prod/dev)"`
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

	// =========================================================================
	// DB

	db, err := gorm.Open(cfg.DB.Driver, cfg.DB.Database)
	if err != nil {
		sugared.With("error", err).Panic("failed to connect database")
	}

	defer func() {
		err := db.Close()
		if err != nil {
			sugared.With("error", err).Error("error while closing database handler")
		}
	}()

	sugared.Info("performing maintenance script")

	if cfg.DryRun {
		sugared.Info("running in dry-run mode")
	} else {
		sugared.Infof("applying changes in non-dry-run mode in %s", cfg.Delay)
		time.Sleep(cfg.Delay)
	}

	if !db.HasTable(internal.Role{}) {
		sugared.With("table", "role").Info("adding missing table")
		if !cfg.DryRun {
			db.CreateTable(internal.Role{})
		}
	}
	
	return nil
}
