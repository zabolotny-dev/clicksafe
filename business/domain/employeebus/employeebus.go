package employeebus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

var (
	ErrUniqueEmail = errors.New("employee with this email already exists")
	ErrUniquePhone = errors.New("employee with this phone already exists")
	ErrNotFound    = errors.New("employee not found")
)

type Storer interface {
	Save(ctx context.Context, employee Employee) (Employee, error)
	QueryByID(ctx context.Context, id uuid.UUID) (Employee, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Employee, error)
	Update(ctx context.Context, employee Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type Business struct {
	storer        Storer
	departmentBus departmentbus.ExtBusiness
}

func NewBusiness(storer Storer, departmentBus departmentbus.ExtBusiness) *Business {
	return &Business{storer: storer, departmentBus: departmentBus}
}

func (b *Business) Save(ctx context.Context, employee NewEmployee) (Employee, error) {
	if employee.DepartmentID != nil {
		_, err := b.departmentBus.QueryByID(ctx, *employee.DepartmentID)
		if err != nil {
			return Employee{}, fmt.Errorf("department.querybyid: %s: %w", *employee.DepartmentID, err)
		}
	}

	emp := Employee{
		ID:           uuid.New(),
		DepartmentID: employee.DepartmentID,
		FirstName:    employee.FirstName,
		LastName:     employee.LastName,
		Email:        employee.Email,
		Phone:        employee.Phone,
		Attributes:   employee.Attributes,
	}

	if emp, err := b.storer.Save(ctx, emp); err != nil {
		return Employee{}, fmt.Errorf("save: %w", err)
	} else {
		return emp, nil
	}
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Employee, error) {
	dep, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Employee{}, fmt.Errorf("query: employeeID[%s]: %w", id, err)
	}

	return dep, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Employee, error) {
	emps, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return emps, nil
}

func (b *Business) Update(ctx context.Context, emp Employee, up UpdateEmployee) error {
	if up.DepartmentID != nil {
		if _, err := b.departmentBus.QueryByID(ctx, *up.DepartmentID); err != nil {
			return fmt.Errorf("department.querybyid: %s: %w", *up.DepartmentID, err)
		}
		emp.DepartmentID = up.DepartmentID
	}

	if up.FirstName != nil {
		emp.FirstName = *up.FirstName
	}

	if up.LastName != nil {
		emp.LastName = *up.LastName
	}

	if up.Email != nil {
		emp.Email = *up.Email
	}

	if up.Phone != nil {
		emp.Phone = *up.Phone
	}

	if up.Attributes != nil {
		emp.Attributes = *up.Attributes
	}

	if err := b.storer.Update(ctx, emp); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

func (b *Business) Delete(ctx context.Context, id uuid.UUID) error {
	if err := b.storer.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}
