package employeeapp

import (
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
)

var orderByFields = map[string]string{
	"employee_id": employeebus.OrderByID,
	"first_name":  employeebus.OrderByFirstName,
	"last_name":   employeebus.OrderByLastName,
	"email":       employeebus.OrderByEmail,
	"phone":       employeebus.OrderByPhone,
}
