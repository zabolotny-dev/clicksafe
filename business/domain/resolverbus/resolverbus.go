package resolverbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

var (
	ErrTargetIDRequired     = errors.New("target id is required")
	ErrTargetNotFound       = errors.New("target not found")
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUnsupportedPath      = errors.New("unsupported resolver path")
	ErrDomainRequired       = errors.New("domain not found")
)

type targetQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error)
}

type targetLinkResolver interface {
	PhishingURL(ctx context.Context, id uuid.UUID) (string, error)
}

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
	targetQuerier   targetQuerier
	targetLinkBus   targetLinkResolver
	employeeBus     employeeQuerier
	departmentBus   departmentQuerier
	organizationBus organizationGetter
}

func NewBusiness(targetQuerier targetQuerier, targetLinkBus targetLinkResolver, employeeBus employeeQuerier, departmentBus departmentQuerier, organizationBus organizationGetter) *Business {
	return &Business{
		targetQuerier:   targetQuerier,
		targetLinkBus:   targetLinkBus,
		employeeBus:     employeeBus,
		departmentBus:   departmentBus,
		organizationBus: organizationBus,
	}
}

func (b *Business) Resolve(ctx context.Context, scope Scope, paths []string) (Result, error) {
	if len(paths) == 0 {
		return Result{Data: map[string]any{}}, nil
	}

	if scope.TargetID == uuid.Nil {
		return Result{}, ErrTargetIDRequired
	}

	target, err := b.targetQuerier.QueryByID(ctx, scope.TargetID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrTargetNotFound) {
			return Result{}, ErrTargetNotFound
		}

		return Result{}, fmt.Errorf("resolve target: %w", err)
	}

	state := resolveState{
		target:          target,
		employeeBus:     b.employeeBus,
		departmentBus:   b.departmentBus,
		organizationBus: b.organizationBus,
		targetLinkBus:   b.targetLinkBus,
	}
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
