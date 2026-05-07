package targetapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, campaignbus.ErrTargetNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, campaignbus.ErrEmployeeNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, campaignbus.ErrCampaignNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, campaignbus.ErrTargetAlreadyExists):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, campaignbus.ErrUniqueToken):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, campaignbus.ErrInvalidStatusTransition):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrTargetLocked):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrInvalidScheduleWindow):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, campaignbus.ErrCampaignLocked):
		return errs.New(errs.FailedPrecondition, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
