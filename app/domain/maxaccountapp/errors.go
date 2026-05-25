package maxaccountapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, maxaccountbus.ErrAccountNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, maxaccountbus.ErrAdapterFailed):
		return errs.New(errs.Unavailable, err)
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
