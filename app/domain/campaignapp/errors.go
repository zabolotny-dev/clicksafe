package campaignapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

func mapBusErr(err error, op string) error {
	switch {
	case errors.Is(err, campaignbus.ErrNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, campaignbus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", op, err)
	}
}
