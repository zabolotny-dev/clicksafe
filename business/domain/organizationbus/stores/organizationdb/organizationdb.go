package organizationdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb/sqlc"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, organization organizationbus.Organization) error {
	org, err := toDBOrganization(organization)
	if err != nil {
		return err
	}

	return s.q.Save(ctx, sqlc.SaveParams{
		ID:         org.ID,
		Name:       org.Name,
		LogoUrl:    org.LogoUrl,
		Attributes: org.Attributes,
	})
}

func (s *Store) Get(ctx context.Context, id uuid.UUID) (organizationbus.Organization, error) {
	org, err := s.q.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organizationbus.Organization{}, organizationbus.ErrNotFound
		}
		return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
	}

	return toBusOrganization(org)
}
