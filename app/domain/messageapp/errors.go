package messageapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

func mapBusErr(err error, msg string) error {

	switch {
	case errors.Is(err, messagebus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, messagebus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, messagebus.ErrInvalidAttachment):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
