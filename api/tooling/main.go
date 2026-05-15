package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/zabolotny-dev/clicksafe/api/tooling/commands"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/foundation/password"
)

func main() {
	if err := run(); err != nil {
		if !errors.Is(err, commands.ErrHelp) {
			fmt.Println("msg", err)
		}
		os.Exit(1)
	}
}

func run() error {
	cfg := struct {
		Args conf.Args
		DB   struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:vladick,mask"`
			Host         string `conf:"default:localhost:5432"`
			Name         string `conf:"default:clicksafe"`
			MaxOpenConns int    `conf:"default:5"`
		}
		Migration struct {
			Timeout time.Duration `conf:"default:10s"`
		}
		Auth struct {
			ArgonMemory      uint32 `conf:"default:65536"`
			ArgonIterations  uint32 `conf:"default:3"`
			ArgonParallelism uint8  `conf:"default:2"`
			ArgonSaltLength  uint32 `conf:"default:16"`
			ArgonKeyLength   uint32 `conf:"default:32"`
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
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
	}

	hasherConfig := password.Argon2idConfig{
		Memory:      cfg.Auth.ArgonMemory,
		Iterations:  cfg.Auth.ArgonIterations,
		Parallelism: cfg.Auth.ArgonParallelism,
		SaltLength:  cfg.Auth.ArgonSaltLength,
		KeyLength:   cfg.Auth.ArgonKeyLength,
	}

	return processCommands(cfg.Args, cfg.Migration.Timeout, dbConfig, hasherConfig)
}

func processCommands(args conf.Args, timeOut time.Duration, dbConfig database.Config, hasherConfig password.Argon2idConfig) error {
	switch args.Num(0) {
	case "migrate", "up":
		return commands.Migrate(dbConfig, timeOut)

	case "rollback", "down":
		return commands.Rollback(dbConfig, timeOut)

	case "status":
		return commands.Status(dbConfig, timeOut)

	case "reset":
		return commands.Reset(dbConfig, timeOut)

	case "adminadd":
		firstname := args.Num(1)
		lastname := args.Num(2)
		login := args.Num(3)
		password := args.Num(4)
		return commands.UserAdd(dbConfig, firstname, lastname, login, password, timeOut, hasherConfig)

	case "admins":
		login := args.Num(1)
		fullname := args.Num(2)
		return commands.Admins(dbConfig, login, fullname, timeOut)

	case "revoke":
		adminID := args.Num(1)
		return commands.RevokeSessions(dbConfig, adminID, timeOut)

	default:
		fmt.Println("migrate/up:    create the schema in the database")
		fmt.Println("rollback/down: roll back the most recent migration")
		fmt.Println("status:        print the status of all migrations")
		fmt.Println("reset:         roll back all migrations")
		fmt.Println("adminadd:      add a new admin (adminadd <first> <last> <login> <pass>)")
		fmt.Println("admins:        list admins (admins [login_filter] [fullname_filter])")
		fmt.Println("revoke:        revoke all sessions for an admin (revoke <admin_id>)")

		return commands.ErrHelp
	}
}
