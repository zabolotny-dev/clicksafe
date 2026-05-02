package departmentbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type QueryFilter struct {
	ID   *uuid.UUID
	Name *label.Label
}
