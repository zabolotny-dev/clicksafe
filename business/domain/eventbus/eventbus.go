package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Storer interface {
	Save(ctx context.Context, event Event) error
}

type Business struct {
	storer Storer
}

func NewBusinnes(storer Storer) *Business {
	b := Business{
		storer: storer,
	}

	return &b
}

func (b *Business) Publish(ctx context.Context, event Event) error {
	event.ID = uuid.New()
	event.OccurredAt = time.Now().UTC()

	if err := b.storer.Save(ctx, event); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
