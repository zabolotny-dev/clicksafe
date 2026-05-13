package landingapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Landing struct {
	ID           uuid.UUID     `json:"id"`
	Label        string        `json:"label"`
	AttachmentID uuid.NullUUID `json:"attachment_id"`
}

type NewLanding struct {
	Label        string `json:"label"`
	AttachmentID string `json:"attachment_id"`
}

type UpdateLanding struct {
	Label        *string `json:"label"`
	AttachmentID *string `json:"attachment_id"`
}

func toBusNewLanding(req NewLanding) (landingbus.NewLanding, error) {
	var fieldErrors errs.FieldErrors

	lbl, err := label.Parse(req.Label)
	if err != nil {
		fieldErrors.Add("label", err)
	}

	var attachmentID uuid.NullUUID
	if req.AttachmentID != "" {
		parsed, err := uuid.Parse(req.AttachmentID)
		if err != nil {
			fieldErrors.Add("attachment_id", err)
		}
		attachmentID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	if len(fieldErrors) > 0 {
		return landingbus.NewLanding{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return landingbus.NewLanding{
		Label:        lbl,
		AttachmentID: attachmentID,
	}, nil
}

func toBusUpdateLanding(req UpdateLanding) (landingbus.UpdateLanding, error) {
	var fieldErrors errs.FieldErrors

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			fieldErrors.Add("label", err)
		}
		lbl = &parsed
	}

	var attachmentID *uuid.NullUUID
	if req.AttachmentID != nil {
		attachmentID = &uuid.NullUUID{}
		if id := *req.AttachmentID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				fieldErrors.Add("attachment_id", err)
			}
			attachmentID.UUID = parsed
			attachmentID.Valid = err == nil
		}
	}

	if len(fieldErrors) > 0 {
		return landingbus.UpdateLanding{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return landingbus.UpdateLanding{
		Label:        lbl,
		AttachmentID: attachmentID,
	}, nil
}

func toAppLanding(landing landingbus.Landing) Landing {
	return Landing{
		ID:           landing.ID,
		Label:        landing.Label.String(),
		AttachmentID: landing.AttachmentID,
	}
}

func toAppLandings(landings []landingbus.Landing) []Landing {
	items := make([]Landing, len(landings))
	for i, landing := range landings {
		items[i] = toAppLanding(landing)
	}

	return items
}
