package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/migrate"
)

var ErrHelp = errors.New("provided help")

func Migrate(cfg database.Config, timeOut time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	pgx, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer pgx.Close()

	db := stdlib.OpenDBFromPool(pgx)
	defer db.Close()

	fmt.Println("Applying migrations...")

	if err := migrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	fmt.Println("Migrations applied successfully")

	return nil
}
