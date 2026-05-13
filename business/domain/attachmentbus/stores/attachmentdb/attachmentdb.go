package attachmentdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus/stores/attachmentdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

const uniqueLabelConstraint = "attachments_label_key"

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, atch attachmentbus.Attachment) error {
	dbAtch := toDBAttachment(atch)

	err := s.q.Save(ctx, sqlc.SaveParams{
		ID:           dbAtch.ID,
		Label:        dbAtch.Label,
		Type:         dbAtch.Type,
		ContentPath:  dbAtch.ContentPath,
		RequiredVars: dbAtch.RequiredVars,
		IsPublic:     dbAtch.IsPublic,
		UploadedAt:   dbAtch.UploadedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return attachmentbus.ErrUniqueLabel
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, atch attachmentbus.Attachment) error {
	dbAtch := toDBAttachment(atch)

	err := s.q.Update(ctx, sqlc.UpdateParams{
		ID:           dbAtch.ID,
		Label:        dbAtch.Label,
		Type:         dbAtch.Type,
		ContentPath:  dbAtch.ContentPath,
		RequiredVars: dbAtch.RequiredVars,
		IsPublic:     dbAtch.IsPublic,
		UploadedAt:   dbAtch.UploadedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return attachmentbus.ErrUniqueLabel
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, atch attachmentbus.Attachment) error {
	if err := s.q.Delete(ctx, atch.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.FKViolation {
			return attachmentbus.ErrInUse
		}

		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	dbAtch, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
		}

		return attachmentbus.Attachment{}, fmt.Errorf("db: %w", err)
	}

	atch, err := toBusAttachment(dbAtch)
	if err != nil {
		return attachmentbus.Attachment{}, fmt.Errorf("db: %w", err)
	}

	return atch, nil
}

func (s *Store) Query(ctx context.Context, filter attachmentbus.QueryFilter, orderBy order.By, page page.Page) ([]attachmentbus.Attachment, error) {
	dbFilter := toDBFilter(filter)

	dbAttachments, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:        filter.ID,
		Label:     dbFilter.Label,
		Type:      dbFilter.Type,
		OrderBy:   orderBy.SQLOrderBy(),
		OffsetVal: int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:  int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	attachments, err := toBusAttachments(dbAttachments)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return attachments, nil
}

func (s *Store) Count(ctx context.Context, filter attachmentbus.QueryFilter) (int, error) {
	dbFilter := toDBFilter(filter)

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:    filter.ID,
		Label: dbFilter.Label,
		Type:  dbFilter.Type,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(count), nil
}
