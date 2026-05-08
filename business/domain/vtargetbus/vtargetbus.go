package vtargetbus

import (
	"context"
	"fmt"

	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type Storer interface {
	Query(ctx context.Context, filter Filter, orderBy order.By, page page.Page) ([]Target, error)
	Count(ctx context.Context, filter Filter) (int, error)
}

type Business struct {
	storer Storer
}

func NewBusiness(storer Storer) *Business {
	return &Business{storer: storer}
}

func (b *Business) Query(ctx context.Context, filter Filter, orderBy order.By, page page.Page) ([]Target, error) {
	targets, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return targets, nil
}

func (b *Business) Count(ctx context.Context, filter Filter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
