package departmentbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Department struct {
	ID         uuid.UUID
	Label      label.Label
	Attributes map[string]string
}

type NewDepartment struct {
	Label      label.Label
	Attributes map[string]string
}

type UpdateDepartment struct {
	Label      *label.Label
	Attributes *map[string]string
}
