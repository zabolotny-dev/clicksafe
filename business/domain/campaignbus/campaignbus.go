package campaignbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

var (
	ErrNotFound    = errors.New("campaign not found")
	ErrUniqueLabel = errors.New("campaign with this label already exists")
)

type Storer interface {
	Save(ctx context.Context, campaign Campaign) error
	Update(ctx context.Context, campaign Campaign) error
	QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Campaign, error)
	Delete(ctx context.Context, campaign Campaign) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type Business struct {
	storer      Storer
	messagesBus messagebus.ExtBusiness
}

func NewBusiness(storer Storer, messagesBus messagebus.ExtBusiness) *Business {
	return &Business{storer: storer, messagesBus: messagesBus}
}

func (b *Business) Save(ctx context.Context, campaign NewCampaign) (Campaign, error) {
	cmp := Campaign{
		ID:         uuid.New(),
		MessageID:  campaign.MessageID,
		Label:      campaign.Label,
		Status:     Draft,
		DateRange:  campaign.DateRange,
		Attributes: campaign.Attributes,
	}

	if err := b.storer.Save(ctx, cmp); err != nil {
		return Campaign{}, fmt.Errorf("save: %w", err)
	}

	return cmp, nil
}

func (b *Business) Update(ctx context.Context, cmp Campaign, upd UpdateCampaign) (Campaign, error) {
	if upd.MessageID != nil {
		if _, err := b.messagesBus.QueryByID(ctx, *upd.MessageID); err != nil {
			return Campaign{}, fmt.Errorf("update.querybyid %s: %w", *upd.MessageID, err)
		}

		cmp.MessageID = upd.MessageID
	}

	if upd.Label != nil {
		cmp.Label = *upd.Label
	}

	if upd.DateRange != nil {
		cmp.DateRange = *upd.DateRange
	}

	if upd.Attributes != nil {
		cmp.Attributes = *upd.Attributes
	}

	if err := b.storer.Update(ctx, cmp); err != nil {
		return Campaign{}, fmt.Errorf("update: %w", err)
	}

	return cmp, nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	cmp, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("query: campaignID[%s]: %w", id, err)
	}

	return cmp, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Campaign, error) {
	camps, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return camps, nil
}

func (b *Business) Delete(ctx context.Context, campaign Campaign) error {
	if err := b.storer.Delete(ctx, campaign); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}
