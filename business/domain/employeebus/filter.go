package employeebus

import (
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID           *uuid.UUID
	DepartmentID *uuid.UUID
	FullName     *string
	Email        *string
	Phone        *string
}
