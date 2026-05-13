package landingbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Storer interface {
	Save(ctx context.Context, landing Landing) error
	Update(ctx context.Context, landing Landing) error
	Delete(ctx context.Context, landing Landing) error
	QueryByID(ctx context.Context, id uuid.UUID) (Landing, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Landing, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type AttachmentQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type Business struct {
	storer            Storer
	attachmentQuerier AttachmentQuerier
}

func NewBusiness(storer Storer, attachmentQuerier AttachmentQuerier) *Business {
	return &Business{storer: storer, attachmentQuerier: attachmentQuerier}
}

func (b *Business) Save(ctx context.Context, landing NewLanding) (Landing, error) {
	l := Landing{
		ID:    uuid.New(),
		Label: landing.Label,
	}

	if landing.AttachmentID.Valid {
		atch, err := b.attachmentQuerier.QueryByID(ctx, landing.AttachmentID.UUID)
		if err != nil {
			return Landing{}, fmt.Errorf("save: %w", err)
		}

		if atch.Type != attachmentbus.Html {
			return Landing{}, fmt.Errorf("save: %w", ErrInvalidAttachment)
		}

		l.AttachmentID = landing.AttachmentID
	}

	if err := b.storer.Save(ctx, l); err != nil {
		return Landing{}, fmt.Errorf("save: %w", err)
	}

	return l, nil
}

func (b *Business) Update(ctx context.Context, landing Landing, up UpdateLanding) (Landing, error) {
	if up.Label != nil {
		landing.Label = *up.Label
	}

	if up.AttachmentID != nil {
		if up.AttachmentID.Valid {
			atch, err := b.attachmentQuerier.QueryByID(ctx, up.AttachmentID.UUID)
			if err != nil {
				return Landing{}, fmt.Errorf("update: %w", err)
			}

			if atch.Type != attachmentbus.Html {
				return Landing{}, fmt.Errorf("update: %w", ErrInvalidAttachment)
			}
		}

		landing.AttachmentID = *up.AttachmentID
	}

	if err := b.storer.Update(ctx, landing); err != nil {
		return Landing{}, fmt.Errorf("update: %w", err)
	}

	return landing, nil
}

func (b *Business) Delete(ctx context.Context, landing Landing) error {
	if err := b.storer.Delete(ctx, landing); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Landing, error) {
	l, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Landing{}, fmt.Errorf("query: landingID[%s]: %w", id, err)
	}

	return l, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Landing, error) {
	landings, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return landings, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
