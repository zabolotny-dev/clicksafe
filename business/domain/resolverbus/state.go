package resolverbus

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

type resolveState struct {
	employee employeebus.Employee

	departmentBus departmentQuerier
	department    departmentbus.Department
	hasDepartment bool
	depLoaded     bool

	organizationBus organizationGetter
	organization    organizationbus.Organization
	orgLoaded       bool
}

func (s *resolveState) resolvePath(ctx context.Context, path string) (string, bool, error) {
	segments, err := splitPath(path)
	if err != nil {
		return "", false, err
	}

	rootValue, isMissing, err := s.rootValue(ctx, segments[0])
	if err != nil || isMissing {
		return "", isMissing, err
	}

	current := rootValue

	for _, segment := range segments[1:] {
		current, isMissing = walkSegment(current, segment)
		if isMissing {
			return "", true, nil
		}
	}

	value, isMissing, err := stringifyValue(current)
	if err != nil || isMissing {
		return "", isMissing, err
	}

	return value, false, nil
}

func (s *resolveState) rootValue(ctx context.Context, root string) (reflect.Value, bool, error) {
	switch root {
	case rootEmployee:
		return reflect.ValueOf(s.employee), false, nil

	case rootDepartment:
		department, isMissing, err := s.loadDepartment(ctx)
		if err != nil || isMissing {
			return reflect.Value{}, isMissing, err
		}

		return reflect.ValueOf(department), false, nil

	case rootOrganization:
		organization, err := s.loadOrganization(ctx)
		if err != nil {
			return reflect.Value{}, false, err
		}

		return reflect.ValueOf(organization), false, nil

	default:
		return reflect.Value{}, false, fmt.Errorf("%w: %s", ErrUnsupportedPath, root)
	}
}

func (s *resolveState) loadDepartment(ctx context.Context) (departmentbus.Department, bool, error) {
	if s.depLoaded {
		if !s.hasDepartment {
			return departmentbus.Department{}, true, nil
		}

		return s.department, false, nil
	}

	s.depLoaded = true

	if s.employee.DepartmentID == nil {
		return departmentbus.Department{}, true, nil
	}

	department, err := s.departmentBus.QueryByID(ctx, *s.employee.DepartmentID)
	if err != nil {
		if errors.Is(err, departmentbus.ErrNotFound) {
			return departmentbus.Department{}, false, fmt.Errorf("resolve department: %w", ErrDepartmentNotFound)
		}

		return departmentbus.Department{}, false, fmt.Errorf("resolve department: %w", err)
	}

	s.department = department
	s.hasDepartment = true

	return s.department, false, nil
}

func (s *resolveState) loadOrganization(ctx context.Context) (organizationbus.Organization, error) {
	if s.orgLoaded {
		return s.organization, nil
	}

	s.orgLoaded = true

	organization, err := s.organizationBus.Get(ctx)
	if err != nil {
		if errors.Is(err, organizationbus.ErrNotFound) {
			return organizationbus.Organization{}, fmt.Errorf("resolve organization: %w", ErrOrganizationNotFound)
		}

		return organizationbus.Organization{}, fmt.Errorf("resolve organization: %w", err)
	}

	s.organization = organization

	return s.organization, nil
}
