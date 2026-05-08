package resolverbus

import (
	"context"
	"errors"
	"net/mail"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	departmentID := uuid.New()
	targetID := uuid.New()
	employeeID := uuid.New()

	employeeBus := &employeeQuerierStub{employee: testEmployee(employeeID, &departmentID)}
	departmentBus := &departmentQuerierStub{department: testDepartment(departmentID)}
	organizationBus := &organizationGetterStub{organization: testOrganization()}
	targetQBus := &targetQuerierStub{target: testTarget(targetID, employeeID)}

	business := NewBusiness(targetQBus, &targetLinkResolverStub{}, employeeBus, departmentBus, organizationBus)

	result, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{
		"Employee.FirstName",
		"Employee.Email.Address",
		"Employee.Attributes.Bebra",
		"Department.Label",
		"Organization.Attributes.SupportEmail",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{
		Data: map[string]any{
			"Employee": map[string]any{
				"FirstName": "Ivan",
				"Email": map[string]any{
					"Address": "ivan@example.com",
				},
				"Attributes": map[string]any{
					"Bebra": "123",
				},
			},
			"Department": map[string]any{
				"Label": "Human Resources",
			},
			"Organization": map[string]any{
				"Attributes": map[string]any{
					"SupportEmail": "support@clicksafe.test",
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}

	if employeeBus.calls != 1 {
		t.Fatalf("employee bus calls = %d, want 1", employeeBus.calls)
	}

	if departmentBus.calls != 1 {
		t.Fatalf("department bus calls = %d, want 1", departmentBus.calls)
	}

	if organizationBus.calls != 1 {
		t.Fatalf("organization bus calls = %d, want 1", organizationBus.calls)
	}
}

func TestResolveMissingPathsPreserveOrderAndDedupe(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()

	employeeBus := &employeeQuerierStub{employee: testEmployee(employeeID, nil)}
	targetQBus := &targetQuerierStub{target: testTarget(targetID, employeeID)}

	business := NewBusiness(targetQBus, &targetLinkResolverStub{}, employeeBus, &departmentQuerierStub{}, &organizationGetterStub{})

	result, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{
		"Employee.Name",
		"Employee.Attributes.Unknown",
		"Department.Label",
		"Employee.Name",
		"Employee.Attributes.Unknown",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expectedMissing := []string{"Employee.Name", "Employee.Attributes.Unknown", "Department.Label"}
	if !reflect.DeepEqual(result.Missing, expectedMissing) {
		t.Fatalf("Resolve missing = %v, want %v", result.Missing, expectedMissing)
	}

	if !reflect.DeepEqual(result.Data, map[string]any{}) {
		t.Fatalf("Resolve data = %#v, want empty map", result.Data)
	}

	if employeeBus.calls != 1 {
		t.Fatalf("employee bus calls = %d, want 1", employeeBus.calls)
	}
}

func TestResolveEmptyPathsDoesNotRequireTarget(t *testing.T) {
	t.Parallel()

	targetQBus := &targetQuerierStub{err: errors.New("should not be called")}
	business := NewBusiness(targetQBus, &targetLinkResolverStub{}, &employeeQuerierStub{}, &departmentQuerierStub{}, &organizationGetterStub{})

	result, err := business.Resolve(context.Background(), Scope{}, nil)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{Data: map[string]any{}}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}

	if targetQBus.calls != 0 {
		t.Fatalf("target querier calls = %d, want 0", targetQBus.calls)
	}
}

func TestResolveRequiresTargetID(t *testing.T) {
	t.Parallel()

	business := NewBusiness(&targetQuerierStub{}, &targetLinkResolverStub{}, &employeeQuerierStub{}, &departmentQuerierStub{}, &organizationGetterStub{})

	_, err := business.Resolve(context.Background(), Scope{}, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrTargetIDRequired) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrTargetIDRequired)
	}
}

func TestResolveMapsTargetNotFound(t *testing.T) {
	t.Parallel()

	business := NewBusiness(
		&targetQuerierStub{err: campaignbus.ErrTargetNotFound},
		&targetLinkResolverStub{},
		&employeeQuerierStub{},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{TargetID: uuid.New()}, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrTargetNotFound)
	}
}

func TestResolveMapsEmployeeNotFound(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, uuid.New())},
		&targetLinkResolverStub{},
		&employeeQuerierStub{err: employeebus.ErrNotFound},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrEmployeeNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrEmployeeNotFound)
	}
}

func TestResolveRejectsInvalidPaths(t *testing.T) {
	t.Parallel()

	paths := []string{"", " ", "Employee", "Employee.", "Employee..FirstName", ".Employee.FirstName"}

	targetID := uuid.New()
	employeeID := uuid.New()

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			business := NewBusiness(
				&targetQuerierStub{target: testTarget(targetID, employeeID)},
				&targetLinkResolverStub{},
				&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
				&departmentQuerierStub{},
				&organizationGetterStub{},
			)

			_, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{path})
			if !errors.Is(err, ErrUnsupportedPath) {
				t.Fatalf("Resolve error = %v, want %v", err, ErrUnsupportedPath)
			}
		})
	}
}

func TestResolveRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()

	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		&targetLinkResolverStub{},
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{"Campaign.Label"})
	if !errors.Is(err, ErrUnsupportedPath) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrUnsupportedPath)
	}
}

func TestResolveReturnsDepartmentLookupError(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	departmentID := uuid.New()

	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		&targetLinkResolverStub{},
		&employeeQuerierStub{employee: testEmployee(employeeID, &departmentID)},
		&departmentQuerierStub{err: departmentbus.ErrNotFound},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{"Department.Label"})
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}

	if errors.Is(err, departmentbus.ErrNotFound) {
		t.Fatalf("Resolve leaked departmentbus error: %v", err)
	}

	if !errors.Is(err, ErrDepartmentNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrDepartmentNotFound)
	}
}

func TestResolveReturnsOrganizationLookupError(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()

	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		&targetLinkResolverStub{},
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{err: organizationbus.ErrNotFound},
	)

	_, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{"Organization.Label"})
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}

	if errors.Is(err, organizationbus.ErrNotFound) {
		t.Fatalf("Resolve leaked organizationbus error: %v", err)
	}

	if !errors.Is(err, ErrOrganizationNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrOrganizationNotFound)
	}
}

func TestResolveTreatsNullOptionalLeafAsMissing(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()

	employee := testEmployee(employeeID, nil)
	employee.Phone = phone.Null{}
	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		&targetLinkResolverStub{},
		&employeeQuerierStub{employee: employee},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	result, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{"Employee.Phone"})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{
		Data:    map[string]any{},
		Missing: []string{"Employee.Phone"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}
}

func TestResolveTargetLink(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	targetStub := &targetLinkResolverStub{url: "https://phishing.example.com/abc-123"}

	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		targetStub,
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	result, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{
		"Target.Link",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{
		Data: map[string]any{
			"Target": map[string]any{
				"Link": "https://phishing.example.com/abc-123",
			},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}

	if targetStub.calls != 1 {
		t.Fatalf("target stub calls = %d, want 1", targetStub.calls)
	}
}

func TestResolveTargetLinkCachesResult(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	targetStub := &targetLinkResolverStub{url: "https://phishing.example.com/abc-123"}

	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, employeeID)},
		targetStub,
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	// Resolve Target.Link twice in same call — should only call PhishingURL once
	result, err := business.Resolve(context.Background(), Scope{TargetID: targetID}, []string{
		"Target.Link",
		"Target.Link",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{
		Data: map[string]any{
			"Target": map[string]any{
				"Link": "https://phishing.example.com/abc-123",
			},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}

	if targetStub.calls != 1 {
		t.Fatalf("target stub calls = %d, want 1 (should be cached)", targetStub.calls)
	}
}

// =============================================================================
// Test Helpers

func testTarget(id, employeeID uuid.UUID) campaignbus.Target {
	return campaignbus.Target{
		ID:         id,
		Token:      "test-token",
		EmployeeID: employeeID,
		CampaignID: uuid.New(),
		Status:     campaignbus.Pending,
	}
}

func testEmployee(id uuid.UUID, departmentID *uuid.UUID) employeebus.Employee {
	phoneNumber, err := phone.ParseNull("+79991234567")
	if err != nil {
		panic(err)
	}

	return employeebus.Employee{
		ID:           id,
		DepartmentID: departmentID,
		FirstName:    name.MustParse("Ivan"),
		LastName:     name.MustParse("Ivanov"),
		Email:        mail.Address{Address: "ivan@example.com"},
		Phone:        phoneNumber,
		Attributes: map[string]string{
			"Bebra": "123",
		},
	}
}

func testDepartment(id uuid.UUID) departmentbus.Department {
	return departmentbus.Department{
		ID:    id,
		Label: label.MustParse("Human Resources"),
		Attributes: map[string]string{
			"Chat": "hr-chat",
		},
	}
}

func testOrganization() organizationbus.Organization {
	return organizationbus.Organization{
		ID:    uuid.New(),
		Label: label.MustParse("ClickSafe Organization"),
		Attributes: map[string]string{
			"SupportEmail": "support@clicksafe.test",
		},
	}
}

// =============================================================================
// Stubs

type targetQuerierStub struct {
	target campaignbus.Target
	err    error
	calls  int
}

func (s *targetQuerierStub) QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error) {
	s.calls++

	if s.err != nil {
		return campaignbus.Target{}, s.err
	}

	return s.target, nil
}

type targetLinkResolverStub struct {
	url   string
	err   error
	calls int
}

func (s *targetLinkResolverStub) PhishingURL(ctx context.Context, id uuid.UUID) (string, error) {
	s.calls++

	if s.err != nil {
		return "", s.err
	}

	return s.url, nil
}

type employeeQuerierStub struct {
	employee employeebus.Employee
	err      error
	calls    int
}

func (s *employeeQuerierStub) QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error) {
	s.calls++

	if s.err != nil {
		return employeebus.Employee{}, s.err
	}

	return s.employee, nil
}

type departmentQuerierStub struct {
	department departmentbus.Department
	err        error
	calls      int
}

func (s *departmentQuerierStub) QueryByID(ctx context.Context, id uuid.UUID) (departmentbus.Department, error) {
	s.calls++

	if s.err != nil {
		return departmentbus.Department{}, s.err
	}

	return s.department, nil
}

type organizationGetterStub struct {
	organization organizationbus.Organization
	err          error
	calls        int
}

func (s *organizationGetterStub) Get(ctx context.Context) (organizationbus.Organization, error) {
	s.calls++

	if s.err != nil {
		return organizationbus.Organization{}, s.err
	}

	return s.organization, nil
}
