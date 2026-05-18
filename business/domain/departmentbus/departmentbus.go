package departmentbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

var (
	ErrUniqueLabel = errors.New("department with this label already exists")
	ErrNotFound    = errors.New("department not found")
)

type Storer interface {
	Save(ctx context.Context, department Department) error
	SaveMany(ctx context.Context, departments []Department) error
	Update(ctx context.Context, department Department) error
	Delete(ctx context.Context, department Department) error
	QueryByID(ctx context.Context, id uuid.UUID) (Department, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Department, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type Business struct {
	storer Storer
}

func NewBusiness(s Storer) *Business {
	return &Business{storer: s}
}

func (b *Business) Save(ctx context.Context, department NewDepartment) (Department, error) {
	dep := Department{
		ID:         uuid.New(),
		Label:      department.Label,
		Attributes: department.Attributes,
	}

	if err := b.storer.Save(ctx, dep); err != nil {
		return Department{}, fmt.Errorf("save: %w", err)
	}

	return dep, nil
}

func (b *Business) SaveMany(ctx context.Context, departments []NewDepartment) error {
	deps := make([]Department, len(departments))
	for i, department := range departments {
		deps[i] = Department{
			ID:         uuid.New(),
			Label:      department.Label,
			Attributes: department.Attributes,
		}
	}

	if len(deps) == 0 {
		return nil
	}

	if err := b.storer.SaveMany(ctx, deps); err != nil {
		return fmt.Errorf("savemany: %w", err)
	}

	return nil
}

func (b *Business) Update(ctx context.Context, department Department, up UpdateDepartment) (Department, error) {
	if up.Label != nil {
		department.Label = *up.Label
	}

	if up.Attributes != nil {
		department.Attributes = *up.Attributes
	}

	if err := b.storer.Update(ctx, department); err != nil {
		return Department{}, fmt.Errorf("update: %w", err)
	}

	return department, nil
}

func (b *Business) Delete(ctx context.Context, department Department) error {
	if err := b.storer.Delete(ctx, department); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Department, error) {
	dep, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Department{}, fmt.Errorf("query: departmentID[%s]: %w", id, err)
	}

	return dep, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Department, error) {
	deps, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return deps, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}
