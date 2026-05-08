package campaignapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

func mapBusErr(err error, op string) error {
	var unscheduledTargets *campaignbus.ErrUnscheduledTargets

	switch {
	case errors.Is(err, campaignbus.ErrCampaignNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, campaignbus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, campaignbus.ErrMessageNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, campaignbus.ErrInvalidStatusTransition):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrMessageRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrDateRangeRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrDateRangeExpired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrTargetsRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrCampaignLocked):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrDomainRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrLandingNotFound):
		return errs.New(errs.NotFound, err)

	case errors.As(err, &unscheduledTargets):
		var fe errs.FieldErrors
		for _, t := range unscheduledTargets.TargetIDs {
			fe.Add("target_id", errors.New(t.String()))
		}
		return fe.ToError(errs.FailedPrecondition, unscheduledTargets.Error())

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", op, err)
	}
}
