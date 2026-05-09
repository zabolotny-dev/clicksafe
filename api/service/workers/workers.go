package workers

import (
	"context"
	"errors"
	"time"

	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/deliverybus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
	"github.com/zabolotny-dev/clicksafe/foundation/worker"
)

type Worker struct {
	Log         *logger.Logger
	CampaignBus *campaignbus.CampaignBusiness
	DeliveryBus *deliverybus.Business
	Interval    time.Duration
	tickers     []*worker.Ticker
}

func NewWorker(log *logger.Logger, campaignBus *campaignbus.CampaignBusiness, deliveryBus *deliverybus.Business, time time.Duration) *Worker {
	return &Worker{Log: log, CampaignBus: campaignBus, DeliveryBus: deliveryBus, Interval: time}
}

func (w *Worker) Run(ctx context.Context) {
	emailSender := worker.NewTicker(w.Interval, func(ctx context.Context) {
		errs := w.DeliveryBus.SendMail(ctx)
		for _, err := range errs {
			w.Log.Error(ctx, "delivery", "err", err)
		}
	})

	campaignCompletion := worker.NewTicker(w.Interval, func(ctx context.Context) {
		errs := w.CampaignBus.CompleteExpired(ctx)
		for _, err := range errs {
			w.Log.Error(ctx, "campaign completion", "err", err)
		}
	})

	w.tickers = append(w.tickers, emailSender, campaignCompletion)

	for _, ticker := range w.tickers {
		ticker.Start(ctx)
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
