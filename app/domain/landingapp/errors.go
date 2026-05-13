package landingapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, landingbus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, landingbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, landingbus.ErrInvalidAttachment):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
