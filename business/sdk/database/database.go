package database

import (
	"context"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	UniqueViolation = "23505"
	FKViolation     = "23503"
)

type Config struct {
	Host         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: "sslmode=disable&timezone=utc",
	}
	dsn := u.String()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
