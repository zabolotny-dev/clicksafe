package campaignapp

import (
	"errors"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type targetMissingVarsValue struct {
	TargetID   uuid.UUID `json:"target_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Vars       []string  `json:"vars"`
}

func mapBusErr(err error, op string) error {
	var unscheduledTargets *campaignbus.ErrUnscheduledTargets
	var targetsMissingVars *campaignbus.ErrTargetsMissingVars

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

	case errors.Is(err, campaignbus.ErrLandingRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrMessageHTMLRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, campaignbus.ErrLandingHTMLRequired):
		return errs.New(errs.FailedPrecondition, err)

	case errors.As(err, &unscheduledTargets):
		var fe errs.FieldErrors
		for _, t := range unscheduledTargets.TargetIDs {
			fe.Add("target_id", errors.New(t.String()))
		}
		return fe.ToError(errs.FailedPrecondition, unscheduledTargets.Error())

	case errors.As(err, &targetsMissingVars):
		var fe errs.FieldErrors
		for _, t := range targetsMissingVars.Targets {
			fe.AddValue("target_missing_vars", targetMissingVarsValue{
				TargetID:   t.TargetID,
				EmployeeID: t.EmployeeID,
				Vars:       t.Vars,
			})
		}
		return fe.ToError(errs.FailedPrecondition, targetsMissingVars.Error())

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", op, err)
	}
}
