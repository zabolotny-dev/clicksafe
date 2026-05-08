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
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/targetdb/sqlc"
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

func (s *Store) Save(ctx context.Context, target campaignbus.Target) error {
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
					return campaignbus.ErrUniqueToken
				case uniqueEmployeeCampaignKeyConstraint:
					return campaignbus.ErrTargetAlreadyExists
				}
			case database.FKViolation:
				switch pgErr.ConstraintName {
				case employeeFKConstraint:
					return campaignbus.ErrEmployeeNotFound
				case campaignFKConstraint:
					return campaignbus.ErrCampaignNotFound
				}
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, t campaignbus.Target) error {
	if err := s.q.Delete(ctx, t.ID); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, target campaignbus.Target) error {
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
					return campaignbus.ErrUniqueToken
				case uniqueEmployeeCampaignKeyConstraint:
					return campaignbus.ErrTargetAlreadyExists
				}
			case database.FKViolation:
				switch pgErr.ConstraintName {
				case employeeFKConstraint:
					return campaignbus.ErrEmployeeNotFound
				case campaignFKConstraint:
					return campaignbus.ErrCampaignNotFound
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

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error) {
	target, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaignbus.Target{}, campaignbus.ErrTargetNotFound
		}
		return campaignbus.Target{}, fmt.Errorf("db: %w", err)
	}

	busTarget, err := toBusTarget(target)
	if err != nil {
		return campaignbus.Target{}, fmt.Errorf("db: %w", err)
	}

	return busTarget, nil
}

func (s *Store) QueryByToken(ctx context.Context, token string) (campaignbus.Target, error) {
	target, err := s.q.QueryByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return campaignbus.Target{}, campaignbus.ErrTargetNotFound
		}
		return campaignbus.Target{}, fmt.Errorf("db: %w", err)
	}

	busTarget, err := toBusTarget(target)
	if err != nil {
		return campaignbus.Target{}, fmt.Errorf("db: %w", err)
	}

	return busTarget, nil
}

func (s *Store) Query(ctx context.Context, filter campaignbus.TargetQueryFilter) ([]campaignbus.Target, error) {
	dbFilter := toDBFilter(filter)
	targets, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:          dbFilter.ID,
		CampaignID:  dbFilter.CampaignID,
		EmployeeID:  dbFilter.EmployeeID,
		Status:      dbFilter.Status,
		HasSchedule: dbFilter.HasSchedule,
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return toBusTargets(targets)
}

func (s *Store) Count(ctx context.Context, filter campaignbus.TargetQueryFilter) (int, error) {
	dbFilter := toDBFilter(filter)
	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:          dbFilter.ID,
		CampaignID:  dbFilter.CampaignID,
		EmployeeID:  dbFilter.EmployeeID,
		Status:      dbFilter.Status,
		HasSchedule: dbFilter.HasSchedule,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(count), nil
}

func (s *Store) QueryDue(ctx context.Context, now time.Time) ([]campaignbus.Target, error) {
	targets, err := s.q.QueryDue(ctx, pgtype.Timestamptz{Time: now, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return toBusTargets(targets)
}

func (s *Store) UpdateMany(ctx context.Context, targets []campaignbus.Target) error {
	ids := make([]uuid.UUID, len(targets))
	tokens := make([]string, len(targets))
	scheduledAts := make([]pgtype.Timestamptz, len(targets))
	employeeIds := make([]uuid.UUID, len(targets))
	campaignIds := make([]uuid.UUID, len(targets))
	statuses := make([]string, len(targets))
	createdAts := make([]pgtype.Timestamptz, len(targets))

	for i, t := range targets {
		var sheduledAt pgtype.Timestamptz
		if t.ScheduledAt != nil {
			sheduledAt = pgtype.Timestamptz{Time: *t.ScheduledAt, Valid: true}
		}

		ids[i] = t.ID
		tokens[i] = t.Token
		scheduledAts[i] = sheduledAt
		employeeIds[i] = t.EmployeeID
		campaignIds[i] = t.CampaignID
		statuses[i] = t.Status.String()
		createdAts[i] = pgtype.Timestamptz{Time: t.CreatedAt, Valid: true}
	}

	err := s.q.UpdateMany(ctx, sqlc.UpdateManyParams{
		Ids:          ids,
		Tokens:       tokens,
		EmployeeIds:  employeeIds,
		CampaignIds:  campaignIds,
		Statuses:     statuses,
		ScheduledAts: scheduledAts,
		CreatedAts:   createdAts,
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
