package organizationbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

type Organization struct {
	ID         uuid.UUID
	Name       label.Label
	LogoURL    url.URL
	Attributes map[string]string
}

type NewOrganization struct {
	Name       label.Label
	Attributes map[string]string
}
