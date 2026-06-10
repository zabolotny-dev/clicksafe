package resolverbus

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"

	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

type resolveState struct {
	target campaignbus.Target

	employeeBus employeeQuerier
	employee    employeebus.Employee
	empLoaded   bool

	departmentBus departmentQuerier
	department    departmentbus.Department
	hasDepartment bool
	depLoaded     bool

	organizationBus organizationGetter
	attachmentBus   attachmentQuerier
	organization    organizationData
	orgLoaded       bool

	targetQuerier targetQuerier
	targetLink    string
	linkLoaded    bool

	campaign   campaignbus.Campaign
	campLoaded bool
}

func (s *resolveState) resolvePath(ctx context.Context, path string) (string, bool, error) {
	segments, err := splitPath(path)
	if err != nil {
		return "", false, err
	}

	rootValue, isMissing, err := s.rootValue(ctx, segments)
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

func (s *resolveState) rootValue(ctx context.Context, segments []string) (reflect.Value, bool, error) {
	root := segments[0]

	switch root {
	case rootEmployee:
		employee, err := s.loadEmployee(ctx)
		if err != nil {
			return reflect.Value{}, false, err
		}

		return reflect.ValueOf(employee), false, nil

	case rootDepartment:
		department, isMissing, err := s.loadDepartment(ctx)
		if err != nil || isMissing {
			return reflect.Value{}, isMissing, err
		}

		return reflect.ValueOf(department), false, nil

	case rootOrganization:
		org, err := s.loadOrganization(ctx)
		if err != nil {
			return reflect.Value{}, false, err
		}

		return reflect.ValueOf(org), false, nil

	case rootTarget:
		if segments[1] != "Link" {
			return reflect.ValueOf(targetData{}), false, nil
		}

		targetData, err := s.loadTarget(ctx)
		if err != nil {
			return reflect.Value{}, false, err
		}

		return reflect.ValueOf(targetData), false, nil

	case rootCampaign:
		campaignData, err := s.loadCampaign(ctx)
		if err != nil {
			return reflect.Value{}, false, err
		}

		return reflect.ValueOf(campaignData), false, nil

	default:
		return reflect.Value{}, false, fmt.Errorf("%w: %s", ErrUnsupportedPath, root)
	}
}

func (s *resolveState) loadEmployee(ctx context.Context) (employeebus.Employee, error) {
	if s.empLoaded {
		return s.employee, nil
	}

	s.empLoaded = true

	employee, err := s.employeeBus.QueryByID(ctx, s.target.EmployeeID)
	if err != nil {
		if errors.Is(err, employeebus.ErrNotFound) {
			return employeebus.Employee{}, fmt.Errorf("resolve employee: %w", ErrEmployeeNotFound)
		}

		return employeebus.Employee{}, fmt.Errorf("resolve employee: %w", err)
	}

	s.employee = employee

	return s.employee, nil
}

func (s *resolveState) loadDepartment(ctx context.Context) (departmentbus.Department, bool, error) {
	if s.depLoaded {
		if !s.hasDepartment {
			return departmentbus.Department{}, true, nil
		}

		return s.department, false, nil
	}

	s.depLoaded = true

	employee, err := s.loadEmployee(ctx)
	if err != nil {
		return departmentbus.Department{}, false, err
	}

	if employee.DepartmentID == nil {
		return departmentbus.Department{}, true, nil
	}

	department, err := s.departmentBus.QueryByID(ctx, *employee.DepartmentID)
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

type organizationData struct {
	organizationbus.Organization
	Logo string
}

func (s *resolveState) loadOrganization(ctx context.Context) (organizationData, error) {
	if s.orgLoaded {
		return s.organization, nil
	}

	s.orgLoaded = true

	org, err := s.organizationBus.Get(ctx)
	if err != nil {
		if errors.Is(err, organizationbus.ErrNotFound) {
			return organizationData{}, fmt.Errorf("resolve organization: %w", ErrOrganizationNotFound)
		}

		return organizationData{}, fmt.Errorf("resolve organization: %w", err)
	}

	data := organizationData{Organization: org}
	if org.AttachmentID.Valid {
		attachment, err := s.attachmentBus.QueryByID(ctx, org.AttachmentID.UUID)
		if err == nil {
			p := attachment.ContentPath.String()
			data.Logo = p[:len(p)-len(path.Ext(p))]
		}
	}

	s.organization = data

	return s.organization, nil
}

type targetData struct {
	Link string
}

func (s *resolveState) loadTarget(ctx context.Context) (targetData, error) {
	if s.linkLoaded {
		return targetData{Link: s.targetLink}, nil
	}

	s.linkLoaded = true

	link, err := s.targetQuerier.PhishingURL(ctx, s.target.ID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrDomainRequired) {
			return targetData{}, fmt.Errorf("resolve target link: %w", ErrDomainRequired)
		}
		return targetData{}, fmt.Errorf("resolve target link: %w", err)
	}

	s.targetLink = link

	return targetData{Link: s.targetLink}, nil
}

func (s *resolveState) loadCampaign(ctx context.Context) (campaignbus.Campaign, error) {
	if s.campLoaded {
		return s.campaign, nil
	}

	s.campLoaded = true

	campaign, err := s.targetQuerier.QueryCampaignByID(ctx, s.target.CampaignID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrCampaignNotFound) {
			return campaignbus.Campaign{}, fmt.Errorf("resolve campaign: %w", ErrCampaignNotFound)
		}
		return campaignbus.Campaign{}, fmt.Errorf("resolve campaign: %w", err)
	}

	s.campaign = campaign

	return s.campaign, nil
}
