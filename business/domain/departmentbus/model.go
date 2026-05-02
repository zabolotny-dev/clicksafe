package departmentbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Department struct {
	ID         uuid.UUID
	Name       label.Label
	Attributes map[string]string
}

type NewDepartment struct {
	Name       label.Label
	Attributes map[string]string
}

type UpdateDepartment struct {
	Name       *label.Label
	Attributes *map[string]string
}
