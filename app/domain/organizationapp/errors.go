package organizationapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, organizationbus.ErrNotFound):
		return errs.New(errs.NotFound, err)
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
