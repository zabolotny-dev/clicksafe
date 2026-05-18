package employeebus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

var (
	ErrUniqueEmail        = errors.New("employee with this email already exists")
	ErrUniquePhone        = errors.New("employee with this phone already exists")
	ErrNotFound           = errors.New("employee not found")
	ErrDepartmentNotFound = errors.New("department not found")
)

type Storer interface {
	Save(ctx context.Context, employee Employee) error
	SaveMany(ctx context.Context, employees []Employee) error
	QueryByID(ctx context.Context, id uuid.UUID) (Employee, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Employee, error)
	Update(ctx context.Context, employee Employee) error
	Delete(ctx context.Context, employee Employee) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type Business struct {
	storer Storer
}

func NewBusiness(storer Storer) *Business {
	return &Business{storer: storer}
}

func (b *Business) Save(ctx context.Context, employee NewEmployee) (Employee, error) {
	emp := Employee{
		ID:           uuid.New(),
		DepartmentID: employee.DepartmentID,
		FirstName:    employee.FirstName,
		LastName:     employee.LastName,
		Email:        employee.Email,
		Phone:        employee.Phone,
		Attributes:   employee.Attributes,
	}

	if err := b.storer.Save(ctx, emp); err != nil {
		if errors.Is(err, ErrDepartmentNotFound) && employee.DepartmentID != nil {
			return Employee{}, fmt.Errorf("save: departmentID[%s]: %w", *employee.DepartmentID, err)
		}
		return Employee{}, fmt.Errorf("save: %w", err)
	}

	return emp, nil
}

func (b *Business) SaveMany(ctx context.Context, employees []NewEmployee) error {
	emps := make([]Employee, len(employees))
	for i, employee := range employees {
		emps[i] = Employee{
			ID:           uuid.New(),
			DepartmentID: employee.DepartmentID,
			FirstName:    employee.FirstName,
			LastName:     employee.LastName,
			Email:        employee.Email,
			Phone:        employee.Phone,
			Attributes:   employee.Attributes,
		}
	}

	if len(emps) == 0 {
		return nil
	}

	if err := b.storer.SaveMany(ctx, emps); err != nil {
		return fmt.Errorf("savemany: %w", err)
	}

	return nil
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

func (b *Business) Update(ctx context.Context, emp Employee, up UpdateEmployee) (Employee, error) {
	if up.DepartmentID != nil {
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
		if errors.Is(err, ErrDepartmentNotFound) && up.DepartmentID != nil {
			return Employee{}, fmt.Errorf("update: departmentID[%s]: %w", *up.DepartmentID, err)
		}
		return Employee{}, fmt.Errorf("update: %w", err)
	}

	return emp, nil
}

func (b *Business) Delete(ctx context.Context, emp Employee) error {
	if err := b.storer.Delete(ctx, emp); err != nil {
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
