package organizationbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Organization struct {
	ID           uuid.UUID
	Label        label.Label
	AttachmentID uuid.NullUUID
	Attributes   map[string]string
}

type NewOrganization struct {
	Label        label.Label
	AttachmentID uuid.NullUUID
	Attributes   map[string]string
}

type UpdateOrganization struct {
	Label        *label.Label
	AttachmentID *uuid.NullUUID
	Attributes   *map[string]string
}
