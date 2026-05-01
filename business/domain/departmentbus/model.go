package departmentbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

type Department struct {
	ID         uuid.UUID
	Name       name.Name
	Attributes map[string]string
}

type NewDepartment struct {
	Name       name.Name
	Attributes map[string]string
}

type UpdateDepartment struct {
	Name       *name.Name
	Attributes *map[string]string
}
