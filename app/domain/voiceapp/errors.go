package voiceapp

import (
	"context"
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
)

func mapBusErr(err error, _ string) error {
	switch {
	case errors.Is(err, context.Canceled):
		return errs.New(errs.Canceled, err)
	case errors.Is(err, context.DeadlineExceeded):
		return errs.New(errs.DeadlineExceeded, err)
	default:
		return errs.New(errs.FailedPrecondition, err)
	}
}
