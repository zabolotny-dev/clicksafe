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
	ErrUniqueName = errors.New("department with this name already exists")
	ErrNotFound   = errors.New("department not found")
)

type Storer interface {
	Save(ctx context.Context, department Department) error
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

func (b *Business) Save(ctx context.Context, department NewDepartment) error {
	dep := Department{
		ID:         uuid.New(),
		Name:       department.Name,
		Attributes: department.Attributes,
	}

	if err := b.storer.Save(ctx, dep); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (b *Business) Update(ctx context.Context, department Department, up UpdateDepartment) error {
	if up.Name != nil {
		department.Name = *up.Name
	}

	if up.Attributes != nil {
		department.Attributes = *up.Attributes
	}

	if err := b.storer.Update(ctx, department); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
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
		return Department{}, fmt.Errorf("querybyid: %w", err)
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
