package messageapp

import (
	"errors"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

func mapBusErr(err error, msg string) error {
	var missingAtch *messagebus.ErrMissingAttachments
	var dupeAtch *messagebus.ErrDuplicateAttachments

	switch {
	case errors.Is(err, messagebus.ErrUniqueLabel):
		return errs.New(errs.AlreadyExists, err)

	case errors.Is(err, messagebus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.Is(err, messagebus.ErrInvalidAttachment):
		return errs.New(errs.InvalidArgument, err)

	case errors.Is(err, attachmentbus.ErrNotFound):
		return errs.New(errs.NotFound, err)

	case errors.As(err, &missingAtch):
		var fe errs.FieldErrors
		for _, v := range missingAtch.IDs {
			fe.Add("attachment_ids", errors.New(v.String()))
		}
		return fe.ToError(errs.NotFound, missingAtch.Error())

	case errors.As(err, &dupeAtch):
		var fe errs.FieldErrors
		for _, v := range dupeAtch.IDs {
			fe.Add("attachment_ids", errors.New(v.String()))
		}
		return fe.ToError(errs.AlreadyExists, dupeAtch.Error())

	default:
		return errs.Errorf(errs.InternalOnlyLog, "%s: %s", msg, err)
	}
}
