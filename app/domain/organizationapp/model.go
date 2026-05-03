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
	LogoPath   string            `json:"logo_path"`
	Attributes map[string]string `json:"attributes"`
}

type NewOrganization struct {
	Label      string            `json:"label"`
	Attributes map[string]string `json:"attributes"`
}

type Logo struct {
	Path file.Path `json:"path"`
}

func toBusNewOrganization(org Organization) (organizationbus.NewOrganization, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(org.Label)
	if err != nil {
		errors.Add("label", err)
	}

	if len(errors) > 0 {
		return organizationbus.NewOrganization{}, errors.ToError()
	}

	return organizationbus.NewOrganization{
		Label:      lbl,
		Attributes: org.Attributes,
	}, nil
}

func toAppOrganization(org organizationbus.Organization) Organization {
	return Organization{
		ID:         org.ID,
		Label:      org.Label.String(),
		LogoPath:   org.LogoPath.String(),
		Attributes: org.Attributes,
	}
}
