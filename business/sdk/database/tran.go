package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxRunner abstracts transaction execution so usecases can be tested without a real pool.
type TxRunner func(ctx context.Context, fn func(context.Context) error) error

// Noop is a TxRunner that runs fn directly without a transaction, for use in tests.
var Noop TxRunner = func(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// NewTxRunner returns a TxRunner backed by a real pgxpool.
func NewTxRunner(pool *pgxpool.Pool) TxRunner {
	return func(ctx context.Context, fn func(context.Context) error) error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback(ctx)

		if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
			return err
		}

		return tx.Commit(ctx)
	}
}

type txKey struct{}

func GetTx(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx)
	return tx
}
