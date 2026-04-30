package migrate

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var EmbedMigrations embed.FS

const MigrationsDir = "sql"

func setup() error {
	goose.SetBaseFS(EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db, MigrationsDir); err != nil {
		return err
	}

	return nil
}

func Rollback(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.DownContext(ctx, db, MigrationsDir); err != nil {
		return err
	}

	return nil
}

func Status(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.StatusContext(ctx, db, MigrationsDir); err != nil {
		return err
	}
	return nil
}

func Reset(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.ResetContext(ctx, db, MigrationsDir); err != nil {
		return err
	}

	return nil
}
