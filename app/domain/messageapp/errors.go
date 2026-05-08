package messageapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
)

func mapBusErr(err error, msg string) error {
	var missingVars *messagebus.MissingRequiredVarsError

	switch {
	case errors.Is(err, messagebus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, messagebus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, messagebus.ErrContentNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, messagebus.ErrEmptyContent):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, messagebus.ErrUnsupportedTemplateSyntax):
		return errs.New(errs.InvalidArgument, err)

	case errors.As(err, &missingVars):
		var fe errs.FieldErrors
		for _, v := range missingVars.Vars {
			fe.Add("var", errors.New(v))
		}
		return fe.ToError(errs.FailedPrecondition, missingVars.Error())

	case errors.Is(err, resolverbus.ErrTargetIDRequired):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, resolverbus.ErrTargetNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, resolverbus.ErrEmployeeNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, resolverbus.ErrDepartmentNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, resolverbus.ErrOrganizationNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, resolverbus.ErrUnsupportedPath):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, resolverbus.ErrDomainRequired):
		return errs.New(errs.FailedPrecondition, resolverbus.ErrDomainRequired)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
