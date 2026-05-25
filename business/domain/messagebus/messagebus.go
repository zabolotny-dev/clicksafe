package messagebus

import (
	"context"
	"errors"
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
	msgType := msg.Type
	if msgType == (MessageType{}) {
		msgType = EmailMessage
	}

	message := Message{
		ID:           uuid.New(),
		Type:         msgType,
		Label:        msg.Label,
		FromEmail:    msg.FromEmail,
		FromName:     msg.FromName,
		Subject:      msg.Subject,
		HtmlBodyID:   msg.HtmlBodyID,
		TextBodyID:   msg.TextBodyID,
		MaxAccountID: msg.MaxAccountID,
	}

	if err := b.validateMessage(ctx, message, "save"); err != nil {
		return Message{}, err
	}

	counts := make(map[uuid.UUID]int)
	for _, id := range msg.AttachmentIDs {
		counts[id]++
	}

	var duplicates []uuid.UUID
	for id, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}

	if len(duplicates) > 0 {
		return Message{}, &ErrDuplicateAttachments{IDs: duplicates}
	}

	message.AttachmentIDs = msg.AttachmentIDs
	if err := b.validateMessageAttachments(ctx, message.Type, message.AttachmentIDs, "save"); err != nil {
		return Message{}, err
	}

	if err := b.storer.Save(ctx, message); err != nil {
		return Message{}, fmt.Errorf("save: %w", err)
	}

	return message, nil
}

func (b *Business) Update(ctx context.Context, msg Message, up UpdateMessage) (Message, error) {
	if up.Type != nil {
		msg.Type = *up.Type
	}
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

	if up.HtmlBodyID != nil {
		if up.HtmlBodyID.Valid {
			atch, err := b.attachmentQuerier.QueryByID(ctx, up.HtmlBodyID.UUID)
			if err != nil {
				return Message{}, fmt.Errorf("update: %w", err)
			}

			if atch.Type != attachmentbus.Html {
				return Message{}, fmt.Errorf("update: %w", ErrInvalidAttachment)
			}
		}

		msg.HtmlBodyID = *up.HtmlBodyID
	}

	if up.TextBodyID != nil {
		msg.TextBodyID = *up.TextBodyID
	}

	if up.MaxAccountID != nil {
		msg.MaxAccountID = *up.MaxAccountID
	}

	if err := b.validateMessage(ctx, msg, "update"); err != nil {
		return Message{}, err
	}

	if up.AttachmentIDs != nil {
		counts := make(map[uuid.UUID]int)
		for _, id := range up.AttachmentIDs {
			counts[id]++
		}

		var duplicates []uuid.UUID
		for id, count := range counts {
			if count > 1 {
				duplicates = append(duplicates, id)
			}
		}

		if len(duplicates) > 0 {
			return Message{}, &ErrDuplicateAttachments{IDs: duplicates}
		}

		msg.AttachmentIDs = up.AttachmentIDs
	}

	if err := b.validateMessageAttachments(ctx, msg.Type, msg.AttachmentIDs, "update"); err != nil {
		return Message{}, err
	}

	if err := b.storer.Update(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("update: %w", err)
	}

	return msg, nil
}

func (b *Business) validateMessage(ctx context.Context, msg Message, op string) error {
	switch msg.Type {
	case EmailMessage:
		if msg.FromEmail.Address == "" {
			return fmt.Errorf("%s: %w", op, ErrFromEmailRequired)
		}

		if msg.HtmlBodyID.Valid {
			atch, err := b.attachmentQuerier.QueryByID(ctx, msg.HtmlBodyID.UUID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if atch.Type != attachmentbus.Html {
				return fmt.Errorf("%s: %w", op, ErrInvalidAttachment)
			}
		}

	case MaxMessage:
		if !msg.MaxAccountID.Valid {
			return fmt.Errorf("%s: %w", op, ErrMaxAccountRequired)
		}
		if !msg.TextBodyID.Valid {
			return fmt.Errorf("%s: %w", op, ErrTextBodyRequired)
		}

		atch, err := b.attachmentQuerier.QueryByID(ctx, msg.TextBodyID.UUID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if atch.Type != attachmentbus.Txt {
			return fmt.Errorf("%s: %w", op, ErrInvalidAttachment)
		}

	default:
		return fmt.Errorf("%s: %w", op, ErrInvalidType)
	}

	return nil
}

func (b *Business) validateMessageAttachments(ctx context.Context, msgType MessageType, ids []uuid.UUID, op string) error {
	var missing []uuid.UUID
	for _, id := range ids {
		attachment, err := b.attachmentQuerier.QueryByID(ctx, id)
		if err != nil {
			if errors.Is(err, attachmentbus.ErrNotFound) {
				missing = append(missing, id)
				continue
			}

			return fmt.Errorf("%s: %w", op, err)
		}

		if msgType == MaxMessage && attachment.Type == attachmentbus.Html {
			return fmt.Errorf("%s: %w", op, ErrMaxHTMLAttachment)
		}
	}

	if len(missing) > 0 {
		return &ErrMissingAttachments{IDs: missing}
	}

	return nil
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
