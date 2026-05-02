package mid

import (
	"context"
	"errors"

	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
)

type ctxKey int

const (
	departmentKey ctxKey = iota + 1
	employeeKey
)

func setDepartment(ctx context.Context, d departmentbus.Department) context.Context {
	return context.WithValue(ctx, departmentKey, d)
}

func GetDepartment(ctx context.Context) (departmentbus.Department, error) {
	d, ok := ctx.Value(departmentKey).(departmentbus.Department)
	if !ok {
		return departmentbus.Department{}, errors.New("department not found in context")
	}
	return d, nil
}

func setEmployee(ctx context.Context, e employeebus.Employee) context.Context {
	return context.WithValue(ctx, employeeKey, e)
}

func GetEmployee(ctx context.Context) (employeebus.Employee, error) {
	e, ok := ctx.Value(employeeKey).(employeebus.Employee)
	if !ok {
		return employeebus.Employee{}, errors.New("employee not found in context")
	}
	return e, nil
}
