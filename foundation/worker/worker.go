package worker

import (
	"context"
	"time"
)

type TickFn func(ctx context.Context)

type Ticker struct {
	interval time.Duration
	fn       TickFn

	cancel context.CancelFunc
	done   chan struct{}
}

func NewTicker(interval time.Duration, fn TickFn) *Ticker {
	return &Ticker{
		interval: interval,
		fn:       fn,
		done:     make(chan struct{}),
	}
}

func (t *Ticker) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel

	go t.run(ctx)
}

func (t *Ticker) run(ctx context.Context) {
	defer close(t.done)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			t.fn(ctx)
		}
	}
}

func (t *Ticker) Shutdown(ctx context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}

	select {
	case <-t.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
