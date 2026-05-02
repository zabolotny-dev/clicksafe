package employeeapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, employeebus.ErrUniqueEmail):
		return errs.New(errs.AlreadyExists, err)
	case errors.Is(err, employeebus.ErrUniquePhone):
		return errs.New(errs.AlreadyExists, err)
	case errors.Is(err, departmentbus.ErrNotFound):
		return errs.New(errs.NotFound, err)
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
