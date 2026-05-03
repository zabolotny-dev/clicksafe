package resolverbus

import (
	"github.com/google/uuid"
)

type Scope struct {
	EmployeeID uuid.UUID
}

type Result struct {
	Data    map[string]any
	Missing []string
}
