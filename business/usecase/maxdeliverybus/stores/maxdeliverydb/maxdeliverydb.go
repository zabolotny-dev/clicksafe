package maxdeliverydb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus/stores/maxdeliverydb/sqlc"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(db)}
}

func (s *Store) queries(ctx context.Context) *sqlc.Queries {
	if tx := database.GetTx(ctx); tx != nil {
		return s.q.WithTx(tx)
	}
	return s.q
}

func (s *Store) SaveSent(ctx context.Context, delivery maxdeliverybus.Delivery) error {
	err := s.queries(ctx).SaveSent(ctx, sqlc.SaveSentParams{
		ID:               delivery.ID,
		TargetID:         delivery.TargetID,
		CampaignID:       delivery.CampaignID,
		EmployeeID:       delivery.EmployeeID,
		MaxAccountID:     delivery.MaxAccountID,
		AdapterAccountID: delivery.AdapterAccountID,
		ChatID:           delivery.ChatID,
		MessageID:        delivery.MessageID,
		SentAt:           toTimestamptz(delivery.SentAt),
		CreatedAt:        toTimestamptz(delivery.CreatedAt),
		UpdatedAt:        toTimestamptz(delivery.UpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("max delivery save sent: %w", err)
	}
	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (maxdeliverybus.Delivery, error) {
	dbDel, err := s.queries(ctx).QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
		}
		return maxdeliverybus.Delivery{}, fmt.Errorf("query max delivery by id: %w", err)
	}
	return toBusDelivery(dbDel), nil
}

func (s *Store) QueryByMessage(ctx context.Context, accountID uuid.UUID, chatID string, messageID string) (maxdeliverybus.Delivery, error) {
	dbDel, err := s.queries(ctx).QueryByMessage(ctx, sqlc.QueryByMessageParams{
		AdapterAccountID: accountID,
		ChatID:           chatID,
		MessageID:        messageID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
		}
		return maxdeliverybus.Delivery{}, fmt.Errorf("query max delivery by message: %w", err)
	}
	return toBusDelivery(dbDel), nil
}

func (s *Store) QueryLatestByChat(ctx context.Context, accountID uuid.UUID, chatID string) (maxdeliverybus.Delivery, error) {
	dbDel, err := s.queries(ctx).QueryLatestByChat(ctx, sqlc.QueryLatestByChatParams{
		AdapterAccountID: accountID,
		ChatID:           chatID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
		}
		return maxdeliverybus.Delivery{}, fmt.Errorf("query latest max delivery by chat: %w", err)
	}
	return toBusDelivery(dbDel), nil
}

func (s *Store) QueryLatestUnreadByChat(ctx context.Context, accountID uuid.UUID, chatID string) (maxdeliverybus.Delivery, error) {
	dbDel, err := s.queries(ctx).QueryLatestUnreadByChat(ctx, sqlc.QueryLatestUnreadByChatParams{
		AdapterAccountID: accountID,
		ChatID:           chatID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
		}
		return maxdeliverybus.Delivery{}, fmt.Errorf("query latest unread max delivery by chat: %w", err)
	}
	return toBusDelivery(dbDel), nil
}

func (s *Store) MarkRead(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	res, err := s.queries(ctx).MarkRead(ctx, sqlc.MarkReadParams{
		ID:     id,
		ReadAt: toTimestamptz(at),
	})
	if err != nil {
		return false, fmt.Errorf("max delivery mark read: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

func (s *Store) MarkReplied(ctx context.Context, id uuid.UUID, incomingMessageID string, at time.Time) (bool, error) {
	res, err := s.queries(ctx).MarkReplied(ctx, sqlc.MarkRepliedParams{
		ID:                id,
		RepliedAt:         toTimestamptz(at),
		IncomingMessageID: incomingMessageID,
	})
	if err != nil {
		return false, fmt.Errorf("max delivery mark replied: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

func (s *Store) MarkEducationSent(ctx context.Context, id uuid.UUID, at time.Time) error {
	err := s.queries(ctx).MarkEducationSent(ctx, sqlc.MarkEducationSentParams{
		ID:        id,
		UpdatedAt: toTimestamptz(at),
	})
	if err != nil {
		return fmt.Errorf("max delivery mark education sent: %w", err)
	}
	return nil
}

func (s *Store) IsProcessed(ctx context.Context, seq int64) (bool, error) {
	exists, err := s.queries(ctx).IsProcessed(ctx, seq)
	if err != nil {
		return false, fmt.Errorf("max event processed lookup: %w", err)
	}
	return exists, nil
}

func (s *Store) MarkProcessed(ctx context.Context, seq int64) error {
	err := s.queries(ctx).MarkProcessed(ctx, sqlc.MarkProcessedParams{
		Seq:         seq,
		ProcessedAt: toTimestamptz(time.Now().UTC()),
	})
	if err != nil {
		return fmt.Errorf("max event mark processed: %w", err)
	}
	return nil
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func toBusDelivery(dbDel sqlc.MaxDelivery) maxdeliverybus.Delivery {
	var readAt *time.Time
	if dbDel.ReadAt.Valid {
		readAt = &dbDel.ReadAt.Time
	}

	var repliedAt *time.Time
	if dbDel.RepliedAt.Valid {
		repliedAt = &dbDel.RepliedAt.Time
	}

	var eduSentAt *time.Time
	if dbDel.EducationSentAt.Valid {
		eduSentAt = &dbDel.EducationSentAt.Time
	}

	return maxdeliverybus.Delivery{
		ID:                dbDel.ID,
		TargetID:          dbDel.TargetID,
		CampaignID:        dbDel.CampaignID,
		EmployeeID:        dbDel.EmployeeID,
		MaxAccountID:      dbDel.MaxAccountID,
		AdapterAccountID:  dbDel.AdapterAccountID,
		ChatID:            dbDel.ChatID,
		MessageID:         dbDel.MessageID,
		SentAt:            dbDel.SentAt.Time,
		ReadAt:            readAt,
		RepliedAt:         repliedAt,
		EducationSentAt:   eduSentAt,
		IncomingMessageID: dbDel.IncomingMessageID,
		CreatedAt:         dbDel.CreatedAt.Time,
		UpdatedAt:         dbDel.UpdatedAt.Time,
	}
}

func (s *Store) QueryReply(ctx context.Context, accountID uuid.UUID, chatID string, replyToMessageID string) (maxdeliverybus.Delivery, error) {
	if replyToMessageID != "" {
		delivery, err := s.QueryByMessage(ctx, accountID, chatID, replyToMessageID)
		if err == nil {
			return delivery, nil
		}
		if !errors.Is(err, maxdeliverybus.ErrDeliveryNotFound) {
			return maxdeliverybus.Delivery{}, err
		}
	}

	return s.QueryLatestByChat(ctx, accountID, chatID)
}

