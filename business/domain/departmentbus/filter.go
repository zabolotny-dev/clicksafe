package departmentbus

import (
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID    *uuid.UUID
	Label *string
}
