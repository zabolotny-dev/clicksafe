package attachmentapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
)

func mapBusErr(err error, msg string) error {
	var missingVars *renderbus.MissingRequiredVarsError

	switch {
	case errors.Is(err, attachmentbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, attachmentbus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, attachmentbus.ErrUnsupportedTemplateSyntax):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrInvalidType):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrContentNotFound):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, attachmentbus.ErrEmptyContent):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrInUse):
		return errs.New(errs.FailedPrecondition, err)

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
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, renderbus.ErrContentNotFound):
		return errs.New(errs.FailedPrecondition, err)

	case errors.Is(err, renderbus.ErrInvalidType):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, renderbus.ErrUnsupportedTemplateSyntax):
		return errs.New(errs.InvalidArgument, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
