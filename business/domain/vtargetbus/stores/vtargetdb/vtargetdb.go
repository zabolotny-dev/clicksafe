package vtargetdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus/stores/vtargetdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Query(ctx context.Context, filter vtargetbus.Filter, orderBy order.By, page page.Page) ([]vtargetbus.Target, error) {
	dbFilter := toDBFilter(filter)

	rows, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:         dbFilter.ID,
		CampaignID: dbFilter.CampaignID,
		EmployeeID: dbFilter.EmployeeID,
		Status:     dbFilter.Status,
		FullName:   dbFilter.FullName,
		OrderBy:    orderBy.SQLOrderBy(),
		OffsetVal:  int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:   int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return toBusTargets(rows), nil
}

func (s *Store) Count(ctx context.Context, filter vtargetbus.Filter) (int, error) {
	dbFilter := toDBFilter(filter)

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:         dbFilter.ID,
		CampaignID: dbFilter.CampaignID,
		EmployeeID: dbFilter.EmployeeID,
		Status:     dbFilter.Status,
		FullName:   dbFilter.FullName,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(count), nil
}
