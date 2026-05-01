package departmentdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus/stores/departmentdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, department departmentbus.Department) error {
	deb, err := toDBDepartment(department)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	if err := s.q.Save(ctx, sqlc.SaveParams{
		ID:         deb.ID,
		Name:       deb.Name,
		Attributes: deb.Attributes,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return departmentbus.ErrUniqueName
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, department departmentbus.Department) error {
	deb, err := toDBDepartment(department)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Update(ctx, sqlc.UpdateParams{
		ID:         deb.ID,
		Name:       deb.Name,
		Attributes: deb.Attributes,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return departmentbus.ErrUniqueName
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, department departmentbus.Department) error {
	if err := s.q.Delete(ctx, department.ID); err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (departmentbus.Department, error) {
	deb, err := s.q.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return departmentbus.Department{}, departmentbus.ErrNotFound
		}
		return departmentbus.Department{}, fmt.Errorf("db: %w", err)
	}
	return toBusDepartment(deb)
}

func (s *Store) Query(ctx context.Context, filter departmentbus.QueryFilter, orderBy order.By, page page.Page) ([]departmentbus.Department, error) {
	var nameFilter pgtype.Text
	if filter.Name != nil {
		nameFilter = pgtype.Text{String: filter.Name.String(), Valid: true}
	}

	debs, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:        filter.ID,
		Name:      nameFilter,
		OrderBy:   orderBy.SQLOrderBy(),
		OffsetVal: int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:  int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	var departments []departmentbus.Department
	for _, deb := range debs {
		department, err := toBusDepartment(deb)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}
		departments = append(departments, department)
	}

	return departments, nil
}

func (s *Store) Count(ctx context.Context, filter departmentbus.QueryFilter) (int, error) {
	var nameFilter pgtype.Text
	if filter.Name != nil {
		nameFilter = pgtype.Text{String: filter.Name.String(), Valid: true}
	}

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:   filter.ID,
		Name: nameFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(count), nil
}
