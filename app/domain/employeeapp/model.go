package employeeapp

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

type Employee struct {
	ID           uuid.UUID         `json:"id"`
	DepartmentID *uuid.UUID        `json:"department_id"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Attributes   map[string]string `json:"attributes"`
}

type NewEmployee struct {
	DepartmentID string            `json:"department_id"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Attributes   map[string]string `json:"attributes"`
}

type UpdateEmployee struct {
	DepartmentID *string            `json:"department_id"`
	FirstName    *string            `json:"first_name"`
	LastName     *string            `json:"last_name"`
	Email        *string            `json:"email"`
	Phone        *string            `json:"phone"`
	Attributes   *map[string]string `json:"attributes"`
}

func toBusNewEmployee(req NewEmployee) (employeebus.NewEmployee, error) {
	var errors errs.FieldErrors

	var dID *uuid.UUID
	if req.DepartmentID != "" {
		id, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			errors.Add("department_id", err)
		}
		dID = &id
	}

	firstName, err := name.Parse(req.FirstName)
	if err != nil {
		errors.Add("first_name", err)
	}

	lastName, err := name.Parse(req.LastName)
	if err != nil {
		errors.Add("last_name", err)
	}

	email, err := mail.ParseAddress(req.Email)
	if err != nil {
		errors.Add("email", err)
	}

	ph, err := phone.ParseNull(req.Phone)
	if err != nil {
		errors.Add("phone", err)
	}

	if len(errors) > 0 {
		return employeebus.NewEmployee{}, errors.ToError()
	}

	return employeebus.NewEmployee{
		DepartmentID: dID,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        *email,
		Phone:        ph,
		Attributes:   req.Attributes,
	}, nil
}

func toAppEmployee(emp employeebus.Employee) Employee {
	return Employee{
		ID:           emp.ID,
		DepartmentID: emp.DepartmentID,
		FirstName:    emp.FirstName.String(),
		LastName:     emp.LastName.String(),
		Email:        emp.Email.Address,
		Phone:        emp.Phone.String(),
		Attributes:   emp.Attributes,
	}
}

func toAppEmployees(emps []employeebus.Employee) []Employee {
	result := make([]Employee, len(emps))
	for i, emp := range emps {
		result[i] = toAppEmployee(emp)
	}
	return result
}

func toBusUpdateEmployee(req UpdateEmployee) (employeebus.UpdateEmployee, error) {
	var errors errs.FieldErrors

	var dID *uuid.UUID
	if req.DepartmentID != nil {
		id, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			errors.Add("department_id", err)
		}
		dID = &id
	}

	var firstName *name.Name
	if req.FirstName != nil {
		fn, err := name.Parse(*req.FirstName)
		if err != nil {
			errors.Add("first_name", err)
		}
		firstName = &fn
	}

	var lastName *name.Name
	if req.LastName != nil {
		ln, err := name.Parse(*req.LastName)
		if err != nil {
			errors.Add("last_name", err)
		}
		lastName = &ln
	}

	var email *mail.Address
	if req.Email != nil {
		e, err := mail.ParseAddress(*req.Email)
		if err != nil {
			errors.Add("email", err)
		}
		email = e
	}

	var ph *phone.Null
	if req.Phone != nil {
		p, err := phone.ParseNull(*req.Phone)
		if err != nil {
			errors.Add("phone", err)
		}
		ph = &p
	}

	if len(errors) > 0 {
		return employeebus.UpdateEmployee{}, errors.ToError()
	}

	return employeebus.UpdateEmployee{
		DepartmentID: dID,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Phone:        ph,
		Attributes:   req.Attributes,
	}, nil
}
