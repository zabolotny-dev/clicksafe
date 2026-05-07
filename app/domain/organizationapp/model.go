package organizationapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Organization struct {
	ID         uuid.UUID         `json:"id"`
	Label      string            `json:"label"`
	LogoPath   *string           `json:"logo_path"`
	Attributes map[string]string `json:"attributes"`
}

type NewOrganization struct {
	Label      string            `json:"label"`
	Attributes map[string]string `json:"attributes"`
}

type UpdateOrganization struct {
	Label      *string            `json:"label"`
	Attributes *map[string]string `json:"attributes"`
}

type Logo struct {
	Path file.Path `json:"path"`
}

func toBusNewOrganization(org NewOrganization) (organizationbus.NewOrganization, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(org.Label)
	if err != nil {
		errors.Add("label", err)
	}

	if len(errors) > 0 {
		return organizationbus.NewOrganization{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return organizationbus.NewOrganization{
		Label:      lbl,
		Attributes: org.Attributes,
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

	if len(errors) > 0 {
		return organizationbus.UpdateOrganization{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return organizationbus.UpdateOrganization{
		Label:      lbl,
		Attributes: org.Attributes,
	}, nil
}

func toAppOrganization(org organizationbus.Organization) Organization {
	var logoPath *string
	if org.LogoPath.Valid() {
		path := org.LogoPath.String()
		logoPath = &path
	}

	return Organization{
		ID:         org.ID,
		Label:      org.Label.String(),
		LogoPath:   logoPath,
		Attributes: org.Attributes,
	}
}
