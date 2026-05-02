package organizationapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

type Organization struct {
	ID         uuid.UUID         `json:"id"`
	Name       string            `json:"name"`
	LogoURL    string            `json:"logo_url"`
	Attributes map[string]string `json:"attributes"`
}

type NewOrganization struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

type Logo struct {
	URL url.URL `json:"url"`
}

func toBusNewOrganization(org Organization) (organizationbus.NewOrganization, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(org.Name)
	if err != nil {
		errors.Add("name", err)
	}

	if len(errors) > 0 {
		return organizationbus.NewOrganization{}, errors.ToError()
	}

	return organizationbus.NewOrganization{
		Name:       lbl,
		Attributes: org.Attributes,
	}, nil
}

func toAppOrganization(org organizationbus.Organization) Organization {
	return Organization{
		ID:         org.ID,
		Name:       org.Name.String(),
		LogoURL:    org.LogoURL.String(),
		Attributes: org.Attributes,
	}
}
