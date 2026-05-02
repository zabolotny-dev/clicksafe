package employeedb

import (
	"encoding/json"
	"net/mail"

	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus/stores/employeedb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

func toDBEmployee(e employeebus.Employee) (sqlc.Employee, error) {
	attributes, err := json.Marshal(e.Attributes)
	if err != nil {
		return sqlc.Employee{}, err
	}

	return sqlc.Employee{
		ID:           e.ID,
		DepartmentID: e.DepartmentID,
		FirstName:    e.FirstName.String(),
		LastName:     e.LastName.String(),
		Email:        e.Email.String(),
		PhoneNumber:  phone.ToSQLNullString(e.Phone),
		Attributes:   attributes,
	}, nil
}

func toBusEmployee(e sqlc.Employee) (employeebus.Employee, error) {
	var attributes map[string]string
	if len(e.Attributes) > 0 {
		if err := json.Unmarshal(e.Attributes, &attributes); err != nil {
			return employeebus.Employee{}, err
		}
	}

	fName, err := name.Parse(e.FirstName)
	if err != nil {
		return employeebus.Employee{}, err
	}

	lName, err := name.Parse(e.LastName)
	if err != nil {
		return employeebus.Employee{}, err
	}

	email, err := mail.ParseAddress(e.Email)
	if err != nil {
		return employeebus.Employee{}, err
	}

	ph, err := phone.ParseNull(e.PhoneNumber.String)
	if err != nil {
		return employeebus.Employee{}, err
	}

	return employeebus.Employee{
		ID:           e.ID,
		DepartmentID: e.DepartmentID,
		FirstName:    fName,
		LastName:     lName,
		Email:        *email,
		Phone:        ph,
		Attributes:   attributes,
	}, nil
}

func toBusEmployees(employees []sqlc.Employee) ([]employeebus.Employee, error) {
	busEmployees := make([]employeebus.Employee, len(employees))

	for i, e := range employees {
		var err error
		busEmployees[i], err = toBusEmployee(e)
		if err != nil {
			return nil, err
		}
	}

	return busEmployees, nil
}
