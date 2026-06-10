package resolverbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
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
	ErrCampaignNotFound     = errors.New("campaign not found")
	ErrUnsupportedPath      = errors.New("unsupported resolver path")
	ErrDomainRequired       = errors.New("domain not found")
)

type targetQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error)
	PhishingURL(ctx context.Context, id uuid.UUID) (string, error)
	QueryCampaignByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error)
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

type attachmentQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type Business struct {
	targetQuerier   targetQuerier
	employeeBus     employeeQuerier
	departmentBus   departmentQuerier
	organizationBus organizationGetter
	attachmentBus   attachmentQuerier
}

func NewBusiness(targetQuerier targetQuerier, employeeBus employeeQuerier, departmentBus departmentQuerier, organizationBus organizationGetter, attachmentBus attachmentQuerier) *Business {
	return &Business{
		targetQuerier:   targetQuerier,
		employeeBus:     employeeBus,
		departmentBus:   departmentBus,
		organizationBus: organizationBus,
		attachmentBus:   attachmentBus,
	}
}

func (b *Business) Resolve(ctx context.Context, targerID uuid.UUID, paths []string) (data map[string]any, missing []string, err error) {
	if len(paths) == 0 {
		return nil, nil, nil
	}

	if targerID == uuid.Nil {
		return nil, nil, ErrTargetIDRequired
	}

	target, err := b.targetQuerier.QueryByID(ctx, targerID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrTargetNotFound) {
			return nil, nil, ErrTargetNotFound
		}

		return nil, nil, fmt.Errorf("resolve target: %w", err)
	}

	state := resolveState{
		target:          target,
		employeeBus:     b.employeeBus,
		departmentBus:   b.departmentBus,
		organizationBus: b.organizationBus,
		attachmentBus:   b.attachmentBus,
		targetQuerier:   b.targetQuerier,
	}
	data = make(map[string]any)
	missingSeen := make(map[string]struct{}, len(paths))

	for _, path := range paths {
		value, isMissing, err := state.resolvePath(ctx, path)
		if err != nil {
			return nil, nil, err
		}

		if isMissing {
			if _, exists := missingSeen[path]; !exists {
				missingSeen[path] = struct{}{}
				missing = append(missing, path)
			}

			continue
		}

		if err := insertResolvedValue(data, path, value); err != nil {
			return nil, nil, fmt.Errorf("resolve path %q: %w", path, err)
		}
	}

	return data, missing, nil
}
