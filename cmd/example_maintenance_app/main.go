package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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

		Logging struct {
			Type string `conf:"default:prod,help:set logging format (prod/dev)"`
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
	var err error

	if cfg.Logging.Type != "dev" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return errors.Wrap(err, "could not initialize logger")
	}
	sugared := logger.Sugar().With("appname", AppName)

	sugared.Info("performing maintenance script")
	
	return nil
}
