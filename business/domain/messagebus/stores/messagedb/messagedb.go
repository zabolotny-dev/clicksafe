package messagedb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

const (
	uniqueLabelConstraint  = "messages_label_key"
	attachmentFKConstraint = "messages_html_body_id_fkey"
	textBodyFKConstraint   = "messages_text_body_id_fkey"
	maxAccountFKConstraint = "messages_max_account_id_fkey"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, msg messagebus.Message) error {
	dbMsg := toDBMessage(msg)

	err := s.q.Save(ctx, sqlc.SaveParams{
		ID:           dbMsg.ID,
		Type:         dbMsg.Type,
		Label:        dbMsg.Label,
		FromEmail:    dbMsg.FromEmail,
		FromName:     dbMsg.FromName,
		Subject:      dbMsg.Subject,
		HtmlBodyID:   dbMsg.HtmlBodyID,
		TextBodyID:   dbMsg.TextBodyID,
		MaxAccountID: dbMsg.MaxAccountID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return messagebus.ErrUniqueLabel
			}
		}
		if errors.As(err, &pgErr) && pgErr.Code == database.FKViolation {
			if pgErr.ConstraintName == attachmentFKConstraint {
				return messagebus.ErrInvalidAttachment
			}
			if pgErr.ConstraintName == textBodyFKConstraint {
				return messagebus.ErrInvalidAttachment
			}
			if pgErr.ConstraintName == maxAccountFKConstraint {
				return messagebus.ErrMaxAccountRequired
			}
		}

		return fmt.Errorf("db: %w", err)
	}

	err = s.q.SyncAttachments(ctx, sqlc.SyncAttachmentsParams{
		MessageID:     msg.ID,
		AttachmentIds: msg.AttachmentIDs,
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, msg messagebus.Message) error {
	dbMsg := toDBMessage(msg)

	err := s.q.Update(ctx, sqlc.UpdateParams{
		ID:           dbMsg.ID,
		Type:         dbMsg.Type,
		Label:        dbMsg.Label,
		FromEmail:    dbMsg.FromEmail,
		FromName:     dbMsg.FromName,
		Subject:      dbMsg.Subject,
		HtmlBodyID:   dbMsg.HtmlBodyID,
		TextBodyID:   dbMsg.TextBodyID,
		MaxAccountID: dbMsg.MaxAccountID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			if pgErr.ConstraintName == uniqueLabelConstraint {
				return messagebus.ErrUniqueLabel
			}
		}
		if errors.As(err, &pgErr) && pgErr.Code == database.FKViolation {
			if pgErr.ConstraintName == attachmentFKConstraint {
				return messagebus.ErrInvalidAttachment
			}
			if pgErr.ConstraintName == textBodyFKConstraint {
				return messagebus.ErrInvalidAttachment
			}
			if pgErr.ConstraintName == maxAccountFKConstraint {
				return messagebus.ErrMaxAccountRequired
			}
		}

		return fmt.Errorf("db: %w", err)
	}

	err = s.q.SyncAttachments(ctx, sqlc.SyncAttachmentsParams{
		MessageID:     msg.ID,
		AttachmentIds: msg.AttachmentIDs,
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, msg messagebus.Message) error {
	if err := s.q.Delete(ctx, msg.ID); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (messagebus.Message, error) {
	dbMsg, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return messagebus.Message{}, messagebus.ErrNotFound
		}

		return messagebus.Message{}, fmt.Errorf("db: %w", err)
	}

	attachmentIDs, err := s.q.QueryAttachments(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return messagebus.Message{}, fmt.Errorf("db: %w", err)
	}

	msg, err := toBusMessage(dbMsg, attachmentIDs)
	if err != nil {
		return messagebus.Message{}, fmt.Errorf("db: %w", err)
	}

	return msg, nil
}

func (s *Store) Query(ctx context.Context, filter messagebus.QueryFilter, orderBy order.By, page page.Page) ([]messagebus.Message, error) {
	dbFilter := toDBFilter(filter)

	dbMessages, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:        filter.ID,
		Type:      dbFilter.Type,
		Label:     dbFilter.Label,
		FromEmail: dbFilter.FromEmail,
		FromName:  dbFilter.FromName,
		Subject:   dbFilter.Subject,
		OrderBy:   orderBy.SQLOrderBy(),
		OffsetVal: int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:  int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	messageIDs := make([]uuid.UUID, len(dbMessages))
	for i, m := range dbMessages {
		messageIDs[i] = m.ID
	}

	rows, err := s.q.QueryAttachmentsByMessageIDs(ctx, messageIDs)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("db: %w", err)
	}

	attachmentsByMsg := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		attachmentsByMsg[row.MessageID] = append(attachmentsByMsg[row.MessageID], row.AttachmentID)
	}

	messages, err := toBusMessages(dbMessages, attachmentsByMsg)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return messages, nil
}

func (s *Store) Count(ctx context.Context, filter messagebus.QueryFilter) (int, error) {
	dbFilter := toDBFilter(filter)

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:        filter.ID,
		Type:      dbFilter.Type,
		Label:     dbFilter.Label,
		FromEmail: dbFilter.FromEmail,
		FromName:  dbFilter.FromName,
		Subject:   dbFilter.Subject,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(count), nil
}
