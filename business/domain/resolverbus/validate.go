package resolverbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

func (b *Business) Validate(ctx context.Context, campaign campaignbus.Campaign, targets []campaignbus.Target, requiredVars []string) ([]campaignbus.TargetMissingVars, error) {
	if len(targets) == 0 || len(requiredVars) == 0 {
		return nil, nil
	}

	needsTargetLink, err := pathsUseRoot(requiredVars, rootTarget)
	if err != nil {
		return nil, fmt.Errorf("validate paths: %w", err)
	}

	if needsTargetLink && campaign.Domain.IsEmpty() {
		return nil, ErrDomainRequired
	}

	departmentBus := newCachedDepartmentQuerier(b.departmentBus)
	organizationBus := &cachedOrganizationGetter{next: b.organizationBus}

	missingByTarget := make([]campaignbus.TargetMissingVars, 0)

	for _, target := range targets {
		state := resolveState{
			target:          target,
			employeeBus:     b.employeeBus,
			departmentBus:   departmentBus,
			organizationBus: organizationBus,
			targetQuerier:   b.targetQuerier,
		}

		if needsTargetLink {
			state.targetLink = campaign.Domain.String() + "/" + target.Token
			state.linkLoaded = true
		}

		missing, err := missingPaths(ctx, &state, requiredVars)
		if err != nil {
			return nil, fmt.Errorf("validate target[%s]: %w", target.ID, err)
		}

		if len(missing) > 0 {
			missingByTarget = append(missingByTarget, campaignbus.TargetMissingVars{
				TargetID:   target.ID,
				EmployeeID: target.EmployeeID,
				Vars:       missing,
			})
		}
	}

	return missingByTarget, nil
}

func pathsUseRoot(paths []string, root string) (bool, error) {
	for _, path := range paths {
		segments, err := splitPath(path)
		if err != nil {
			return false, err
		}

		if segments[0] == root {
			return true, nil
		}
	}

	return false, nil
}

func missingPaths(ctx context.Context, state *resolveState, paths []string) ([]string, error) {
	missingSeen := make(map[string]struct{}, len(paths))
	missing := make([]string, 0)

	for _, path := range paths {
		_, isMissing, err := state.resolvePath(ctx, path)
		if err != nil {
			return nil, err
		}

		if isMissing {
			if _, exists := missingSeen[path]; exists {
				continue
			}

			missingSeen[path] = struct{}{}
			missing = append(missing, path)
		}
	}

	return missing, nil
}

type cachedDepartmentQuerier struct {
	next       departmentQuerier
	department map[uuid.UUID]departmentbus.Department
}

func newCachedDepartmentQuerier(next departmentQuerier) *cachedDepartmentQuerier {
	return &cachedDepartmentQuerier{
		next:       next,
		department: make(map[uuid.UUID]departmentbus.Department),
	}
}

func (c *cachedDepartmentQuerier) QueryByID(ctx context.Context, id uuid.UUID) (departmentbus.Department, error) {
	department, ok := c.department[id]
	if ok {
		return department, nil
	}

	department, err := c.next.QueryByID(ctx, id)
	if err != nil {
		return departmentbus.Department{}, err
	}

	c.department[id] = department

	return department, nil
}

type cachedOrganizationGetter struct {
	next         organizationGetter
	organization organizationbus.Organization
	loaded       bool
}

func (c *cachedOrganizationGetter) Get(ctx context.Context) (organizationbus.Organization, error) {
	if c.loaded {
		return c.organization, nil
	}

	organization, err := c.next.Get(ctx)
	if err != nil {
		return organizationbus.Organization{}, err
	}

	c.organization = organization
	c.loaded = true

	return c.organization, nil
}
