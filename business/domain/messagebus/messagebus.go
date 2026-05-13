package messagebus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Storer interface {
	Save(ctx context.Context, msg Message) error
	Update(ctx context.Context, msg Message) error
	Delete(ctx context.Context, msg Message) error
	QueryByID(ctx context.Context, id uuid.UUID) (Message, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Message, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type AttachmentQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type Business struct {
	storer            Storer
	attachmentQuerier AttachmentQuerier
}

func NewBusiness(s Storer, attachmentQuerier AttachmentQuerier) *Business {
	return &Business{storer: s, attachmentQuerier: attachmentQuerier}
}

func (b *Business) Save(ctx context.Context, msg NewMessage) (Message, error) {
	message := Message{
		ID:        uuid.New(),
		Label:     msg.Label,
		FromEmail: msg.FromEmail,
		FromName:  msg.FromName,
		Subject:   msg.Subject,
	}

	if msg.AttachmentID.Valid {
		atch, err := b.attachmentQuerier.QueryByID(ctx, msg.AttachmentID.UUID)
		if err != nil {
			return Message{}, fmt.Errorf("save: %w", err)
		}

		if atch.Type != attachmentbus.Html {
			return Message{}, fmt.Errorf("save: %w", ErrInvalidAttachment)
		}

		message.AttachmentID = msg.AttachmentID
	}

	if err := b.storer.Save(ctx, message); err != nil {
		return Message{}, fmt.Errorf("save: %w", err)
	}

	return message, nil
}

func (b *Business) Update(ctx context.Context, msg Message, up UpdateMessage) (Message, error) {
	if up.Label != nil {
		msg.Label = *up.Label
	}
	if up.FromEmail != nil {
		msg.FromEmail = *up.FromEmail
	}
	if up.FromName != nil {
		msg.FromName = *up.FromName
	}
	if up.Subject != nil {
		msg.Subject = *up.Subject
	}

	if up.AttachmentID != nil {
		if up.AttachmentID.Valid {
			atch, err := b.attachmentQuerier.QueryByID(ctx, up.AttachmentID.UUID)
			if err != nil {
				return Message{}, fmt.Errorf("update: %w", err)
			}

			if atch.Type != attachmentbus.Html {
				return Message{}, fmt.Errorf("update: %w", ErrInvalidAttachment)
			}
		}

		msg.AttachmentID = *up.AttachmentID
	}

	if err := b.storer.Update(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("update: %w", err)
	}

	return msg, nil
}

func (b *Business) Delete(ctx context.Context, msg Message) error {
	if err := b.storer.Delete(ctx, msg); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Message, error) {
	msg, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Message{}, fmt.Errorf("query: messageID[%s]: %w", id, err)
	}

	return msg, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Message, error) {
	messages, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return messages, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
