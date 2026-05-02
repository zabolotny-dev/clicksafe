package employeebus

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

type Employee struct {
	ID           uuid.UUID
	DepartmentID *uuid.UUID
	FirstName    name.Name
	LastName     name.Name
	Email        mail.Address
	Phone        phone.Null
	Attributes   map[string]string
}

type NewEmployee struct {
	DepartmentID *uuid.UUID
	FirstName    name.Name
	LastName     name.Name
	Email        mail.Address
	Phone        phone.Null
	Attributes   map[string]string
}

type UpdateEmployee struct {
	DepartmentID *uuid.UUID
	FirstName    *name.Name
	LastName     *name.Name
	Email        *mail.Address
	Phone        *phone.Null
	Attributes   *map[string]string
}
