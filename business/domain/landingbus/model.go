package landingbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Landing struct {
	ID           uuid.UUID
	Label        label.Label
	ContentPath  file.Null
	RequiredVars []string
}

type NewLanding struct {
	Label label.Label
}

type UpdateLanding struct {
	Label *label.Label
}
