package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/zabolotny-dev/clicksafe/api/tooling/commands"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

func main() {
	ctx := context.Context(context.Background())

	log := logger.New(os.Stdout, logger.LevelInfo, "tooling")
	if err := run(); err != nil {
		log.Error(ctx, "startup error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := struct {
		Args conf.Args
		DB   struct {
			User     string `conf:"default:postgres"`
			Password string `conf:"default:vladick,mask"`
			Host     string `conf:"default:localhost:5432"`
			Name     string `conf:"default:clicksafe"`
		}
		Migration struct {
			Timeout time.Duration `conf:"default:10s"`
		}
	}{}

	const prefix = "TOOLING"

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	dbConfig := database.Config{
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Host:     cfg.DB.Host,
		Name:     cfg.DB.Name,
	}

	return processCommands(cfg.Args, cfg.Migration.Timeout, dbConfig)
}

func processCommands(args conf.Args, timeOut time.Duration, dbConfig database.Config) error {
	switch args.Num(0) {
	case "migrate", "up":
		return commands.Migrate(dbConfig, timeOut)

	case "rollback", "down":
		return commands.Rollback(dbConfig, timeOut)

	case "status":
		return commands.Status(dbConfig, timeOut)

	case "reset":
		return commands.Reset(dbConfig, timeOut)

	default:
		fmt.Println("migrate/up:    create the schema in the database")
		fmt.Println("rollback/down: roll back the most recent migration")
		fmt.Println("status:        print the status of all migrations")
		fmt.Println("reset:         roll back all migrations")

		return errors.New("unknown command")
	}
}
