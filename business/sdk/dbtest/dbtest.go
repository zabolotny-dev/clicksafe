// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/sdk/migrate"
	"github.com/zabolotny-dev/clicksafe/foundation/docker"
)

// Database owns state for running and shutting down tests.
type Database struct {
	DB        *pgxpool.Pool
	BusDomain BusDomain
}

// New creates a new test database inside a fresh Docker container, migrates it,
// and returns a fully initialised Database with all business domains wired up.
func New(t *testing.T, testName string) *Database {
	t.Helper()

	image := "postgres:16"
	name := "clicksafetest"
	port := "5432"
	dockerArgs := []string{"-e", "POSTGRES_PASSWORD=postgres"}

	c, err := docker.StartContainer(image, name, port, dockerArgs, nil)
	if err != nil {
		t.Fatalf("Starting database container: %v", err)
	}

	t.Logf("Container : %s\n", c.Name)
	t.Logf("HostPort  : %s\n", c.HostPort)

	masterDSN := fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", c.HostPort)

	masterDB, err := sql.Open("pgx", masterDSN)
	if err != nil {
		t.Fatalf("Opening master DB connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := waitReady(ctx, masterDB); err != nil {
		t.Fatalf("Waiting for database: %v", err)
	}

	dbName := randomDBName()
	t.Logf("Create database: %s\n", dbName)

	if _, err := masterDB.ExecContext(context.Background(), "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("Creating test database %s: %v", dbName, err)
	}

	testDSN := fmt.Sprintf("postgres://postgres:postgres@%s/%s?sslmode=disable", c.HostPort, dbName)

	migDB, err := sql.Open("pgx", testDSN)
	if err != nil {
		t.Fatalf("Opening migration DB: %v", err)
	}

	t.Logf("Migrate database: %s\n", dbName)
	if err := migrate.Migrate(context.Background(), migDB); err != nil {
		t.Logf("Migration logs for %s:\n%s", c.Name, docker.DumpContainerLogs(c.Name))
		t.Fatalf("Migrating: %v", err)
	}
	migDB.Close()

	pool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("Opening pgxpool: %v", err)
	}

	tmpDir := t.TempDir()

	t.Cleanup(func() {
		t.Helper()

		pool.Close()

		t.Logf("Drop database: %s\n", dbName)
		if _, err := masterDB.ExecContext(context.Background(), "DROP DATABASE "+dbName); err != nil {
			t.Logf("Dropping database %s: %v", dbName, err)
		}
		masterDB.Close()
	})

	return &Database{
		DB:        pool,
		BusDomain: newBusDomains(pool, tmpDir),
	}
}

func waitReady(ctx context.Context, db *sql.DB) error {
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func randomDBName() string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
