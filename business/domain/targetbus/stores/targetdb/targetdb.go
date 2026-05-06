package targetdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus/stores/targetdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
)

const (
	uniqueTokenConstraint               = "targets_token_key"
	uniqueEmployeeCampaignKeyConstraint = "targets_employee_id_campaign_id_key"
	employeeFKConstraint                = "targets_employee_id_fkey"
	campaignFKConstraint                = "targets_campaign_id_fkey"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, target targetbus.Target) error {
	t := toDBTarget(target)
	err := s.q.Save(ctx, sqlc.SaveParams{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      t.Status,
		ScheduledAt: t.ScheduledAt,
		CreatedAt:   t.CreatedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case database.UniqueViolation:
				switch pgErr.ConstraintName {
				case uniqueTokenConstraint:
					return targetbus.ErrUniqueToken
				case uniqueEmployeeCampaignKeyConstraint:
					return targetbus.ErrTargetAlreadyExists
				}
			case database.FKViolation:
				switch pgErr.ConstraintName {
				case employeeFKConstraint:
					return targetbus.ErrEmployeeNotFound
				case campaignFKConstraint:
					return targetbus.ErrCampaignNotFound
				}
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, t targetbus.Target) error {
	if err := s.q.Delete(ctx, t.ID); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, target targetbus.Target) error {
	t := toDBTarget(target)

	err := s.q.Update(ctx, sqlc.UpdateParams{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      t.Status,
		ScheduledAt: t.ScheduledAt,
		CreatedAt:   t.CreatedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case database.UniqueViolation:
				switch pgErr.ConstraintName {
				case uniqueTokenConstraint:
					return targetbus.ErrUniqueToken
				case uniqueEmployeeCampaignKeyConstraint:
					return targetbus.ErrTargetAlreadyExists
				}
			case database.FKViolation:
				switch pgErr.ConstraintName {
				case employeeFKConstraint:
					return targetbus.ErrEmployeeNotFound
				case campaignFKConstraint:
					return targetbus.ErrCampaignNotFound
				}
			}
		}
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func (s *Store) DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error {
	if err := s.q.DeleteByCampaignID(ctx, campaignID); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (targetbus.Target, error) {
	target, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return targetbus.Target{}, targetbus.ErrNotFound
		}
		return targetbus.Target{}, fmt.Errorf("db: %w", err)
	}

	busTarget, err := toBusTarget(target)
	if err != nil {
		return targetbus.Target{}, fmt.Errorf("db: %w", err)
	}

	return busTarget, nil
}

func (s *Store) QueryByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]targetbus.Target, error) {
	targets, err := s.q.QueryByCampaignID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return toBusTargets(targets)
}

func (s *Store) QueryDue(ctx context.Context, now time.Time) ([]targetbus.Target, error) {
	targets, err := s.q.QueryDue(ctx, pgtype.Timestamptz{Time: now, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return toBusTargets(targets)
}
