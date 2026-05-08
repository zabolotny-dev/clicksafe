package landingdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus/stores/landingdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

const uniqueLabelConstraint = "landings_label_key"

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, landing landingbus.Landing) error {
	dbLanding := toDBLanding(landing)

	err := s.q.Save(ctx, sqlc.SaveParams{
		ID:           dbLanding.ID,
		Label:        dbLanding.Label,
		ContentPath:  dbLanding.ContentPath,
		RequiredVars: dbLanding.RequiredVars,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return landingbus.ErrUniqueLabel
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, landing landingbus.Landing) error {
	dbLanding := toDBLanding(landing)

	err := s.q.Update(ctx, sqlc.UpdateParams{
		ID:           dbLanding.ID,
		Label:        dbLanding.Label,
		ContentPath:  dbLanding.ContentPath,
		RequiredVars: dbLanding.RequiredVars,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return landingbus.ErrUniqueLabel
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, landing landingbus.Landing) error {
	if err := s.q.Delete(ctx, landing.ID); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (landingbus.Landing, error) {
	dbLanding, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return landingbus.Landing{}, landingbus.ErrNotFound
		}

		return landingbus.Landing{}, fmt.Errorf("db: %w", err)
	}

	landing, err := toBusLanding(dbLanding)
	if err != nil {
		return landingbus.Landing{}, fmt.Errorf("db: %w", err)
	}

	return landing, nil
}

func (s *Store) Query(ctx context.Context, filter landingbus.QueryFilter, orderBy order.By, page page.Page) ([]landingbus.Landing, error) {
	dbFilter := toDBFilter(filter)

	dbLandings, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:        filter.ID,
		Label:     dbFilter.Label,
		OrderBy:   orderBy.SQLOrderBy(),
		OffsetVal: int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:  int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	landings, err := toBusLandings(dbLandings)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return landings, nil
}

func (s *Store) Count(ctx context.Context, filter landingbus.QueryFilter) (int, error) {
	dbFilter := toDBFilter(filter)

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:    filter.ID,
		Label: dbFilter.Label,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(count), nil
}
