package vtargetapp

import "github.com/zabolotny-dev/clicksafe/app/sdk/errs"

func mapBusErr(err error, op string) error {
	return errs.Errorf(errs.InternalOnlyLog, "%s: %s", op, err)
}
