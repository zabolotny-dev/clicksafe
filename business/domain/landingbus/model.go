package landingbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Landing struct {
	ID         uuid.UUID
	Label      label.Label
	HtmlBodyID uuid.NullUUID
}

type NewLanding struct {
	Label      label.Label
	HtmlBodyID uuid.NullUUID
}

type UpdateLanding struct {
	Label      *label.Label
	HtmlBodyID *uuid.NullUUID
}
