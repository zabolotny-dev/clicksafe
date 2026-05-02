package organizationdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
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

	return s.q.Save(ctx, sqlc.SaveParams{
		ID:         dbOrg.ID,
		Name:       dbOrg.Name,
		Attributes: dbOrg.Attributes,
	})
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

func (s *Store) UpdateLogo(ctx context.Context, id uuid.UUID, logoURL url.URL) error {
	err := s.q.UpdateLogo(ctx, sqlc.UpdateLogoParams{
		LogoUrl: pgtype.Text{String: logoURL.String(), Valid: !logoURL.IsEmpty()},
		ID:      id,
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
