package maxaccountdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus/stores/maxaccountdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Upsert(ctx context.Context, account maxaccountbus.Account) (maxaccountbus.Account, error) {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}

	now := time.Now().UTC()
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	account.UpdatedAt = now

	dbAcc, err := s.q.Save(ctx, toSaveParams(account))
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}

	saved, err := toBusAccount(dbAcc)
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}
	return saved, nil
}

func (s *Store) UpdateLabel(ctx context.Context, account maxaccountbus.Account) (maxaccountbus.Account, error) {
	account.UpdatedAt = time.Now().UTC()

	dbAcc, err := s.q.UpdateLabel(ctx, sqlc.UpdateLabelParams{
		ID:        account.ID,
		Label:     account.Label,
		UpdatedAt: pgtype.Timestamptz{Time: account.UpdatedAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
		}
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}

	updated, err := toBusAccount(dbAcc)
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}
	return updated, nil
}

func (s *Store) Query(ctx context.Context, filter maxaccountbus.QueryFilter, page page.Page) ([]maxaccountbus.Account, error) {
	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	var phoneFilter pgtype.Text
	if filter.Phone != nil {
		phoneFilter = pgtype.Text{String: *filter.Phone, Valid: true}
	}

	dbAccs, err := s.q.Query(ctx, sqlc.QueryParams{
		Label:       labelFilter,
		PhoneNumber: phoneFilter,
		OffsetVal:   int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:    int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	accounts, err := toBusAccounts(dbAccs)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return accounts, nil
}

func (s *Store) Count(ctx context.Context, filter maxaccountbus.QueryFilter) (int, error) {
	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	var phoneFilter pgtype.Text
	if filter.Phone != nil {
		phoneFilter = pgtype.Text{String: *filter.Phone, Valid: true}
	}

	count, err := s.q.Count(ctx, sqlc.CountParams{
		Label:       labelFilter,
		PhoneNumber: phoneFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(count), nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (maxaccountbus.Account, error) {
	dbAcc, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
		}
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}

	account, err := toBusAccount(dbAcc)
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}
	return account, nil
}

func (s *Store) QueryByAdapterID(ctx context.Context, id uuid.UUID) (maxaccountbus.Account, error) {
	dbAcc, err := s.q.QueryByAdapterID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
		}
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}

	account, err := toBusAccount(dbAcc)
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("db: %w", err)
	}
	return account, nil
}

func (s *Store) Delete(ctx context.Context, account maxaccountbus.Account) error {
	res, err := s.q.Delete(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	if res.RowsAffected() == 0 {
		return maxaccountbus.ErrAccountNotFound
	}
	return nil
}

func toSaveParams(acc maxaccountbus.Account) sqlc.SaveParams {
	var maxUserID pgtype.Text
	if acc.MaxUserID != "" {
		maxUserID = pgtype.Text{String: acc.MaxUserID, Valid: true}
	}

	var lastError pgtype.Text
	if acc.LastError != "" {
		lastError = pgtype.Text{String: acc.LastError, Valid: true}
	}

	return sqlc.SaveParams{
		ID:          acc.ID,
		AdapterID:   acc.AdapterID,
		PhoneNumber: acc.Phone,
		Label:       acc.Label,
		Status:      acc.Status.String(),
		MaxUserID:   maxUserID,
		LastError:   lastError,
		CreatedAt:   pgtype.Timestamptz{Time: acc.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: acc.UpdatedAt, Valid: true},
	}
}
