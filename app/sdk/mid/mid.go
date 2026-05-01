package mid

import (
	"context"
	"errors"

	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
)

type ctxKey int

const (
	departmentKey ctxKey = iota + 1
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
