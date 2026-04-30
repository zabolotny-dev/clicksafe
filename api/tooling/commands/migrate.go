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

func Migrate(cfg database.Config, timeOut time.Duration) error {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	fmt.Println("Applying migrations...")

	if err := migrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	fmt.Println("Migrations applied successfully")

	return nil
}
