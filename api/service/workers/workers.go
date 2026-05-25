package workers

import (
	"context"
	"errors"
	"time"

	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/deliverybus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
	"github.com/zabolotny-dev/clicksafe/foundation/worker"
)

type Worker struct {
	Log            *logger.Logger
	CampaignBus    *campaignbus.CampaignBusiness
	DeliveryBus    *deliverybus.Business
	MaxDeliveryBus *maxdeliverybus.Business
	SessionBus     *sessionbus.Business
	Interval       time.Duration
	tickers        []*worker.Ticker
}

func NewWorker(log *logger.Logger, campaignBus *campaignbus.CampaignBusiness, deliveryBus *deliverybus.Business, maxDeliveryBus *maxdeliverybus.Business, sessionBus *sessionbus.Business, time time.Duration) *Worker {
	return &Worker{Log: log, CampaignBus: campaignBus, DeliveryBus: deliveryBus, MaxDeliveryBus: maxDeliveryBus, SessionBus: sessionBus, Interval: time}
}

func (w *Worker) Run(ctx context.Context) {
	emailSender := worker.NewTicker(w.Interval, func(ctx context.Context) {
		errs := w.DeliveryBus.SendMail(ctx)
		for _, err := range errs {
			w.Log.Error(ctx, "delivery", "err", err)
		}
	})

	maxSender := worker.NewTicker(w.Interval, func(ctx context.Context) {
		errs := w.MaxDeliveryBus.SendDue(ctx)
		for _, err := range errs {
			w.Log.Error(ctx, "max delivery", "err", err)
		}
	})

	go w.runMaxEventConsumer(ctx)

	campaignCompletion := worker.NewTicker(w.Interval, func(ctx context.Context) {
		errs := w.CampaignBus.CompleteExpired(ctx)
		for _, err := range errs {
			w.Log.Error(ctx, "campaign completion", "err", err)
		}
	})

	sessionCleanup := worker.NewTicker(w.Interval, func(ctx context.Context) {
		err := w.SessionBus.DeleteExpired(ctx)
		if err != nil {
			w.Log.Error(ctx, "session cleanup", "err", err)
		}
	})

	w.tickers = append(w.tickers, emailSender, maxSender, campaignCompletion, sessionCleanup)

	for _, ticker := range w.tickers {
		ticker.Start(ctx)
	}
}

func (w *Worker) runMaxEventConsumer(ctx context.Context) {
	for {
		if err := w.MaxDeliveryBus.ConsumeEvents(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			w.Log.Error(ctx, "max events", "err", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.Interval):
		}
	}
}

func (w *Worker) Stop(ctx context.Context) error {
	var errs []error

	for _, ticker := range w.tickers {
		if err := ticker.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
