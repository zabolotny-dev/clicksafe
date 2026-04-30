package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/migrate"
)

func Status(cfg database.Config, timeOut time.Duration) error {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	if err := migrate.Status(ctx, db); err != nil {
		return fmt.Errorf("status: %w", err)
	}

	return nil
}
