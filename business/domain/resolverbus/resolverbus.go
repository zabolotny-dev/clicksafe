package resolverbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

var (
	ErrEmployeeIDRequired   = errors.New("employee id is required")
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUnsupportedPath      = errors.New("unsupported resolver path")
)

type employeeQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error)
}

type departmentQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (departmentbus.Department, error)
}

type organizationGetter interface {
	Get(ctx context.Context) (organizationbus.Organization, error)
}

type ExtBusiness interface {
	Resolve(ctx context.Context, scope Scope, paths []string) (Result, error)
}

type Business struct {
	employeeBus     employeeQuerier
	departmentBus   departmentQuerier
	organizationBus organizationGetter
}

func NewBusiness(employeeBus employeeQuerier, departmentBus departmentQuerier, organizationBus organizationGetter) *Business {
	return &Business{
		employeeBus:     employeeBus,
		departmentBus:   departmentBus,
		organizationBus: organizationBus,
	}
}

func (b *Business) Resolve(ctx context.Context, scope Scope, paths []string) (Result, error) {
	if len(paths) == 0 {
		return Result{Data: map[string]any{}}, nil
	}

	if scope.EmployeeID == uuid.Nil {
		return Result{}, ErrEmployeeIDRequired
	}

	employee, err := b.employeeBus.QueryByID(ctx, scope.EmployeeID)
	if err != nil {
		if errors.Is(err, employeebus.ErrNotFound) {
			return Result{}, ErrEmployeeNotFound
		}

		return Result{}, fmt.Errorf("resolve employee: %w", err)
	}

	state := resolveState{employee: employee, departmentBus: b.departmentBus, organizationBus: b.organizationBus}
	result := Result{Data: make(map[string]any)}
	missingSeen := make(map[string]struct{}, len(paths))

	for _, path := range paths {
		value, isMissing, err := state.resolvePath(ctx, path)
		if err != nil {
			return Result{}, err
		}

		if isMissing {
			if _, exists := missingSeen[path]; !exists {
				missingSeen[path] = struct{}{}
				result.Missing = append(result.Missing, path)
			}

			continue
		}

		if err := insertResolvedValue(result.Data, path, value); err != nil {
			return Result{}, fmt.Errorf("resolve path %q: %w", path, err)
		}
	}

	return result, nil
}
