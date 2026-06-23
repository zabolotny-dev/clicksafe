package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/migrate"
)

func Rollback(cfg database.Config, timeOut time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	pgx, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer pgx.Close()

	db := stdlib.OpenDBFromPool(pgx)
	defer db.Close()

	if err := migrate.Rollback(ctx, db); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}

	fmt.Println("Migration rolled back successfully")

	return nil
}
