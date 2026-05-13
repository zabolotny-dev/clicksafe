package organizationapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Organization struct {
	ID           uuid.UUID         `json:"id"`
	Label        string            `json:"label"`
	AttachmentID uuid.NullUUID     `json:"attachment_id"`
	Attributes   map[string]string `json:"attributes"`
}

type NewOrganization struct {
	Label        string            `json:"label"`
	AttachmentID string            `json:"attachment_id"`
	Attributes   map[string]string `json:"attributes"`
}

type UpdateOrganization struct {
	Label        *string            `json:"label"`
	AttachmentID *string            `json:"attachment_id"`
	Attributes   *map[string]string `json:"attributes"`
}

func toBusNewOrganization(org NewOrganization) (organizationbus.NewOrganization, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(org.Label)
	if err != nil {
		errors.Add("label", err)
	}

	var attachmentID uuid.NullUUID
	if org.AttachmentID != "" {
		parsed, err := uuid.Parse(org.AttachmentID)
		if err != nil {
			errors.Add("attachment_id", err)
		}
		attachmentID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	if len(errors) > 0 {
		return organizationbus.NewOrganization{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return organizationbus.NewOrganization{
		Label:        lbl,
		AttachmentID: attachmentID,
		Attributes:   org.Attributes,
	}, nil
}

func toBusUpdateOrganization(org UpdateOrganization) (organizationbus.UpdateOrganization, error) {
	var errors errs.FieldErrors

	var lbl *label.Label
	if org.Label != nil {
		parsed, err := label.Parse(*org.Label)
		if err != nil {
			errors.Add("label", err)
		}
		lbl = &parsed
	}

	var attachmentID *uuid.NullUUID
	if org.AttachmentID != nil {
		attachmentID = &uuid.NullUUID{}
		if id := *org.AttachmentID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("attachment_id", err)
			}
			attachmentID.UUID = parsed
			attachmentID.Valid = err == nil
		}
	}

	if len(errors) > 0 {
		return organizationbus.UpdateOrganization{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return organizationbus.UpdateOrganization{
		Label:        lbl,
		AttachmentID: attachmentID,
		Attributes:   org.Attributes,
	}, nil
}

func toAppOrganization(org organizationbus.Organization) Organization {
	return Organization{
		ID:           org.ID,
		Label:        org.Label.String(),
		AttachmentID: org.AttachmentID,
		Attributes:   org.Attributes,
	}
}
