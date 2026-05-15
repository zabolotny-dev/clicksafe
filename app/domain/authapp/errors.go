package authapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

func mapBusErr(err error, msg string) error {
	switch {
	case errors.Is(err, adminbus.ErrInvalidCredential):
		return errs.New(errs.Unauthenticated, err)

	case errors.Is(err, adminbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, adminbus.ErrUniqueLogin):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, sessionbus.ErrAdminNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, sessionbus.ErrInvalidCSRF):
		return errs.New(errs.Unauthenticated, err)

	case errors.Is(err, sessionbus.ErrExpired):
		return errs.New(errs.Unauthenticated, err)

	case errors.Is(err, sessionbus.ErrRevoked):
		return errs.New(errs.Unauthenticated, err)

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
