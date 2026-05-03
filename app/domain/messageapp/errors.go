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
		return errs.New(errs.FailedPrecondition, missingVars)
	case errors.Is(err, resolverbus.ErrEmployeeIDRequired):
		return errs.New(errs.InvalidArgument, err)
	case errors.Is(err, resolverbus.ErrEmployeeNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, resolverbus.ErrDepartmentNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, resolverbus.ErrOrganizationNotFound):
		return errs.New(errs.NotFound, err)
	case errors.Is(err, resolverbus.ErrUnsupportedPath):
		return errs.New(errs.InvalidArgument, err)
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
