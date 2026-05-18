package employeeapp

import (
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/csv"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

type csvRowErrorValue struct {
	Row   int    `json:"row"`
	Field string `json:"field,omitempty"`
	Err   string `json:"error"`
}

func csvToBusNewEmployee(row csv.Row) (employeebus.NewEmployee, []csvRowErrorValue) {
	var rowErrs []csvRowErrorValue

	var dID *uuid.UUID
	depIDStr := row.Get("department_id")
	if depIDStr != "" {
		id, err := uuid.Parse(depIDStr)
		if err != nil {
			rowErrs = append(rowErrs, csvRowErrorValue{
				Row:   row.Number(),
				Field: "department_id",
				Err:   fmt.Sprintf("invalid department_id: %s", err),
			})
		} else {
			dID = &id
		}
	}

	var firstName name.Name
	firstName, err := name.Parse(row.Get("first_name"))
	if err != nil {
		rowErrs = append(rowErrs, csvRowErrorValue{
			Row:   row.Number(),
			Field: "first_name",
			Err:   fmt.Sprintf("invalid first_name: %s", err),
		})
	}

	var lastName name.Name
	lastName, err = name.Parse(row.Get("last_name"))
	if err != nil {
		rowErrs = append(rowErrs, csvRowErrorValue{
			Row:   row.Number(),
			Field: "last_name",
			Err:   fmt.Sprintf("invalid last_name: %s", err),
		})
	}

	var email *mail.Address
	email, err = mail.ParseAddress(row.Get("email"))
	if err != nil {
		rowErrs = append(rowErrs, csvRowErrorValue{
			Row:   row.Number(),
			Field: "email",
			Err:   fmt.Sprintf("invalid email: %s", err),
		})
	}

	var ph phone.Null
	phoneStr := row.Get("phone")
	if phoneStr != "" {
		var err error
		ph, err = phone.ParseNull(phoneStr)
		if err != nil {
			rowErrs = append(rowErrs, csvRowErrorValue{
				Row:   row.Number(),
				Field: "phone",
				Err:   fmt.Sprintf("invalid phone: %s", err),
			})
		}
	}

	if len(rowErrs) > 0 {
		return employeebus.NewEmployee{}, rowErrs
	}

	return employeebus.NewEmployee{
		DepartmentID: dID,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        *email,
		Phone:        ph,
		Attributes:   row.Prefixed("attr."),
	}, nil
}
