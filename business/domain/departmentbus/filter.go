package departmentbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}
