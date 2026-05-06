package targetapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, targetbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, targetbus.ErrEmployeeNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, targetbus.ErrCampaignNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, targetbus.ErrTargetAlreadyExists):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, targetbus.ErrUniqueToken):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, targetbus.ErrInvalidStatusValue):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, targetbus.ErrCampaignNotDraft):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, targetbus.ErrInvalidStatusTransition):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, targetbus.ErrTargetNotPending):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, targetbus.ErrInvalidScheduleWindow):
		return errs.New(errs.InvalidArgument, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
