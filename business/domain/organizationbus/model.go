package organizationbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

type Organization struct {
	ID         uuid.UUID
	Name       name.Name
	LogoURL    url.URL
	Attributes map[string]string
}

type NewOrganization struct {
	Name       name.Name
	Attributes map[string]string
}
