package organizationbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Organization struct {
	ID         uuid.UUID
	Label      label.Label
	LogoPath   file.Path
	Attributes map[string]string
}

type NewOrganization struct {
	Label      label.Label
	Attributes map[string]string
}
