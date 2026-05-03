package organizationdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, org organizationbus.Organization) error {
	dbOrg, err := toDBOrganization(org)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Save(ctx, sqlc.SaveParams{
		ID:         dbOrg.ID,
		Label:      dbOrg.Label,
		LogoPath:   dbOrg.LogoPath,
		Attributes: dbOrg.Attributes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			return organizationbus.ErrAlreadyExists
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (organizationbus.Organization, error) {
	dbOrg, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organizationbus.Organization{}, organizationbus.ErrNotFound
		}
		return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
	}

	org, err := toBusOrganization(dbOrg)
	if err != nil {
		return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
	}
	return org, nil
}

func (s *Store) Update(ctx context.Context, org organizationbus.Organization) error {
	dbOrg, err := toDBOrganization(org)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Update(ctx, sqlc.UpdateParams{
		Label:      dbOrg.Label,
		LogoPath:   dbOrg.LogoPath,
		Attributes: dbOrg.Attributes,
		ID:         dbOrg.ID,
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
