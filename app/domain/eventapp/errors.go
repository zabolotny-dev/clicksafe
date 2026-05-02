package eventapp

import (
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
)

func mapBusErr(err error, msg string) error {
	switch {
	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
