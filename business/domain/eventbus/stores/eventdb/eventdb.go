package eventdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus/stores/eventdb/sqlc"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, event eventbus.Event) error {
	err := s.q.SaveEvent(ctx, sqlc.SaveEventParams{
		ID:         event.ID,
		CampaignID: event.CampaignID,
		EmployeeID: event.EmployeeID,
		Type:       event.Type.String(),
		OccurredAt: pgtype.Timestamp{Time: event.OccurredAt, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
