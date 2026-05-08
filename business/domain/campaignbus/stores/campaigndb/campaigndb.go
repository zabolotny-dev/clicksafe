package campaigndb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/campaigndb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

const (
	uniqueLabelConstraint = "campaigns_label_key"
	messageFKConstraint   = "campaigns_message_id_fkey"
	landingFKConstraint   = "campaigns_landing_id_fkey"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, campaign campaignbus.Campaign) error {
	cmp, err := toDBCampaign(campaign)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	err = s.q.Save(ctx, sqlc.SaveParams{
		ID:         cmp.ID,
		MessageID:  cmp.MessageID,
		LandingID:  cmp.LandingID,
		Label:      cmp.Label,
		Domain:     cmp.Domain,
		Status:     cmp.Status,
		DateFrom:   cmp.DateFrom,
		DateTo:     cmp.DateTo,
		Attributes: cmp.Attributes,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case database.UniqueViolation:
				if pgErr.ConstraintName == uniqueLabelConstraint {
					return campaignbus.ErrUniqueLabel
				}
			case database.FKViolation:
				if pgErr.ConstraintName == messageFKConstraint {
					return campaignbus.ErrMessageNotFound
				}
				if pgErr.ConstraintName == landingFKConstraint {
					return campaignbus.ErrLandingNotFound
				}
			}
		}

		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, campaign campaignbus.Campaign) error {
	cmp, err := toDBCampaign(campaign)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Update(ctx, sqlc.UpdateParams{
		MessageID:  cmp.MessageID,
		LandingID:  cmp.LandingID,
		Label:      cmp.Label,
		Domain:     cmp.Domain,
		Status:     cmp.Status,
		DateFrom:   cmp.DateFrom,
		DateTo:     cmp.DateTo,
		Attributes: cmp.Attributes,
		ID:         cmp.ID,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case database.UniqueViolation:
				if pgErr.ConstraintName == uniqueLabelConstraint {
					return campaignbus.ErrUniqueLabel
				}
			case database.FKViolation:
				if pgErr.ConstraintName == messageFKConstraint {
					return campaignbus.ErrMessageNotFound
				}
				if pgErr.ConstraintName == landingFKConstraint {
					return campaignbus.ErrLandingNotFound
				}
			}
		}

		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error) {
	cmp, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaignbus.Campaign{}, campaignbus.ErrCampaignNotFound
		}
		return campaignbus.Campaign{}, fmt.Errorf("db: %w", err)
	}

	buscmp, err := toBusCampaign(cmp)
	if err != nil {
		return campaignbus.Campaign{}, fmt.Errorf("db: %w", err)
	}
	return buscmp, nil
}

func (s *Store) Query(ctx context.Context, filter campaignbus.CampaignQueryFilter, orderBy order.By, page page.Page) ([]campaignbus.Campaign, error) {
	dbFilter := toDBFilter(filter)

	cmps, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:        filter.ID,
		Label:     dbFilter.Label,
		Status:    dbFilter.Status,
		DateFrom:  dbFilter.DateFrom,
		DateTo:    dbFilter.DateTo,
		OrderBy:   orderBy.SQLOrderBy(),
		OffsetVal: int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:  int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	bcmps, err := toBusCampaigns(cmps)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return bcmps, nil
}

func (s *Store) QueryExpired(ctx context.Context) ([]campaignbus.Campaign, error) {
	cmps, err := s.q.QueryExpired(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	bcmps, err := toBusCampaigns(cmps)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return bcmps, nil
}

func (s *Store) Delete(ctx context.Context, campaign campaignbus.Campaign) error {
	err := s.q.Delete(ctx, campaign.ID)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func (s *Store) Count(ctx context.Context, filter campaignbus.CampaignQueryFilter) (int, error) {
	dbFilter := toDBFilter(filter)

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:       filter.ID,
		Label:    dbFilter.Label,
		Status:   dbFilter.Status,
		DateFrom: dbFilter.DateFrom,
		DateTo:   dbFilter.DateTo,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(count), nil
}
