package resolverbus

import (
	"context"
	"errors"
	"net/mail"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
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

	business := NewBusiness(targetQBus, employeeBus, departmentBus, organizationBus, &attachmentQuerierStub{})

	data, missing, err := business.Resolve(context.Background(), targetID, []string{
		"Employee.FirstName",
		"Employee.Email.Address",
		"Employee.Attributes.Bebra",
		"Department.Label",
		"Organization.Attributes.SupportEmail",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expectedData := map[string]any{
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
	}

	if !reflect.DeepEqual(data, expectedData) {
		t.Fatalf("Resolve data = %#v, want %#v", data, expectedData)
	}

	if len(missing) > 0 {
		t.Fatalf("Resolve missing = %v, want empty", missing)
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

	business := NewBusiness(targetQBus, employeeBus, &departmentQuerierStub{}, &organizationGetterStub{}, &attachmentQuerierStub{})

	data, missing, err := business.Resolve(context.Background(), targetID, []string{
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
	if !reflect.DeepEqual(missing, expectedMissing) {
		t.Fatalf("Resolve missing = %v, want %v", missing, expectedMissing)
	}

	if !reflect.DeepEqual(data, map[string]any{}) {
		t.Fatalf("Resolve data = %#v, want empty map", data)
	}

	if employeeBus.calls != 1 {
		t.Fatalf("employee bus calls = %d, want 1", employeeBus.calls)
	}
}

func TestResolveEmptyPathsDoesNotRequireTarget(t *testing.T) {
	t.Parallel()

	targetQBus := &targetQuerierStub{err: errors.New("should not be called")}
	business := NewBusiness(targetQBus, &employeeQuerierStub{}, &departmentQuerierStub{}, &organizationGetterStub{}, &attachmentQuerierStub{})

	data, missing, err := business.Resolve(context.Background(), uuid.Nil, nil)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if data != nil {
		t.Fatalf("Resolve data = %#v, want nil", data)
	}

	if missing != nil {
		t.Fatalf("Resolve missing = %v, want nil", missing)
	}

	if targetQBus.calls != 0 {
		t.Fatalf("target querier calls = %d, want 0", targetQBus.calls)
	}
}

func TestResolveRequiresTargetID(t *testing.T) {
	t.Parallel()

	business := NewBusiness(&targetQuerierStub{}, &employeeQuerierStub{}, &departmentQuerierStub{}, &organizationGetterStub{}, &attachmentQuerierStub{})

	_, _, err := business.Resolve(context.Background(), uuid.Nil, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrTargetIDRequired) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrTargetIDRequired)
	}
}

func TestResolveMapsTargetNotFound(t *testing.T) {
	t.Parallel()

	business := NewBusiness(
		&targetQuerierStub{err: campaignbus.ErrTargetNotFound},
		&employeeQuerierStub{},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	_, _, err := business.Resolve(context.Background(), uuid.New(), []string{"Employee.FirstName"})
	if !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrTargetNotFound)
	}
}

func TestResolveMapsEmployeeNotFound(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	business := NewBusiness(
		&targetQuerierStub{target: testTarget(targetID, uuid.New())},
		&employeeQuerierStub{err: employeebus.ErrNotFound},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	_, _, err := business.Resolve(context.Background(), targetID, []string{"Employee.FirstName"})
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
				&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
				&departmentQuerierStub{},
				&organizationGetterStub{},
				&attachmentQuerierStub{},
			)

			_, _, err := business.Resolve(context.Background(), targetID, []string{path})
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
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	_, _, err := business.Resolve(context.Background(), targetID, []string{"Banana.Label"})
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
		&employeeQuerierStub{employee: testEmployee(employeeID, &departmentID)},
		&departmentQuerierStub{err: departmentbus.ErrNotFound},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	_, _, err := business.Resolve(context.Background(), targetID, []string{"Department.Label"})
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
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{err: organizationbus.ErrNotFound},
		&attachmentQuerierStub{},
	)

	_, _, err := business.Resolve(context.Background(), targetID, []string{"Organization.Label"})
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
		&employeeQuerierStub{employee: employee},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	data, missing, err := business.Resolve(context.Background(), targetID, []string{"Employee.Phone"})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if !reflect.DeepEqual(data, map[string]any{}) {
		t.Fatalf("Resolve data = %#v, want empty map", data)
	}

	expectedMissing := []string{"Employee.Phone"}
	if !reflect.DeepEqual(missing, expectedMissing) {
		t.Fatalf("Resolve missing = %v, want %v", missing, expectedMissing)
	}
}

func TestResolveTargetLink(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	targetStub := &targetQuerierStub{target: testTarget(targetID, employeeID), url: "https://phishing.example.com/abc-123"}

	business := NewBusiness(
		targetStub,
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	data, missing, err := business.Resolve(context.Background(), targetID, []string{
		"Target.Link",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expectedData := map[string]any{
		"Target": map[string]any{
			"Link": "https://phishing.example.com/abc-123",
		},
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Fatalf("Resolve data = %#v, want %#v", data, expectedData)
	}

	if len(missing) > 0 {
		t.Fatalf("Resolve missing = %v, want empty", missing)
	}

	if targetStub.linkCalls != 1 {
		t.Fatalf("target stub calls = %d, want 1", targetStub.calls)
	}
}

func TestResolveTargetLinkCachesResult(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	targetStub := &targetQuerierStub{target: testTarget(targetID, employeeID), url: "https://phishing.example.com/abc-123"}

	business := NewBusiness(
		targetStub,
		&employeeQuerierStub{employee: testEmployee(employeeID, nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	// Resolve Target.Link twice in same call — should only call PhishingURL once
	data, missing, err := business.Resolve(context.Background(), targetID, []string{
		"Target.Link",
		"Target.Link",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expectedData := map[string]any{
		"Target": map[string]any{
			"Link": "https://phishing.example.com/abc-123",
		},
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Fatalf("Resolve data = %#v, want %#v", data, expectedData)
	}

	if len(missing) > 0 {
		t.Fatalf("Resolve missing = %v, want empty", missing)
	}

	if targetStub.linkCalls != 1 {
		t.Fatalf("target stub calls = %d, want 1 (should be cached)", targetStub.calls)
	}
}

func TestValidateReturnsMissingVarsForTargets(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	departmentID := uuid.New()
	firstTargetID := uuid.New()
	secondTargetID := uuid.New()
	firstEmployeeID := uuid.New()
	secondEmployeeID := uuid.New()

	firstEmployee := testEmployee(firstEmployeeID, nil)
	firstEmployee.Phone = phone.Null{}
	secondEmployee := testEmployee(secondEmployeeID, &departmentID)

	targetQuerier := &targetQuerierStub{err: errors.New("should not be called")}
	employeeBus := &employeeQuerierByIDStub{
		employees: map[uuid.UUID]employeebus.Employee{
			firstEmployeeID:  firstEmployee,
			secondEmployeeID: secondEmployee,
		},
	}
	departmentBus := &departmentQuerierStub{department: testDepartment(departmentID)}
	organizationBus := &organizationGetterStub{organization: testOrganization()}

	business := NewBusiness(targetQuerier, employeeBus, departmentBus, organizationBus, &attachmentQuerierStub{})

	targets := []campaignbus.Target{
		{
			ID:         firstTargetID,
			Token:      "first-token",
			EmployeeID: firstEmployeeID,
			CampaignID: campaignID,
			Status:     campaignbus.Pending,
		},
		{
			ID:         secondTargetID,
			Token:      "second-token",
			EmployeeID: secondEmployeeID,
			CampaignID: campaignID,
			Status:     campaignbus.Pending,
		},
	}

	missing, err := business.Validate(context.Background(), campaignbus.Campaign{
		ID:     campaignID,
		Domain: domain.MustParse("https://phishing.example.com"),
	}, targets, []string{
		"Employee.Phone",
		"Department.Label",
		"Organization.Label",
		"Target.Link",
	})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	expected := []campaignbus.TargetMissingVars{
		{
			TargetID:   firstTargetID,
			EmployeeID: firstEmployeeID,
			Vars:       []string{"Employee.Phone", "Department.Label"},
		},
	}
	if !reflect.DeepEqual(missing, expected) {
		t.Fatalf("Validate missing = %#v, want %#v", missing, expected)
	}

	if targetQuerier.calls != 0 {
		t.Fatalf("target querier calls = %d, want 0", targetQuerier.calls)
	}

	if targetQuerier.linkCalls != 0 {
		t.Fatalf("target link calls = %d, want 0", targetQuerier.linkCalls)
	}

	if employeeBus.calls != 2 {
		t.Fatalf("employee bus calls = %d, want 2", employeeBus.calls)
	}

	if departmentBus.calls != 1 {
		t.Fatalf("department bus calls = %d, want 1", departmentBus.calls)
	}

	if organizationBus.calls != 1 {
		t.Fatalf("organization bus calls = %d, want 1", organizationBus.calls)
	}
}

func TestValidateRequiresDomainForTargetLink(t *testing.T) {
	t.Parallel()

	business := NewBusiness(
		&targetQuerierStub{},
		&employeeQuerierStub{},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	_, err := business.Validate(context.Background(), campaignbus.Campaign{}, []campaignbus.Target{
		testTarget(uuid.New(), uuid.New()),
	}, []string{"Target.Link"})
	if !errors.Is(err, ErrDomainRequired) {
		t.Fatalf("Validate error = %v, want %v", err, ErrDomainRequired)
	}
}

func TestValidateUnknownTargetPathDoesNotRequireDomain(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	employeeID := uuid.New()
	targetQuerier := &targetQuerierStub{err: errors.New("should not be called")}
	business := NewBusiness(
		targetQuerier,
		&employeeQuerierStub{},
		&departmentQuerierStub{},
		&organizationGetterStub{},
		&attachmentQuerierStub{},
	)

	missing, err := business.Validate(context.Background(), campaignbus.Campaign{}, []campaignbus.Target{
		testTarget(targetID, employeeID),
	}, []string{"Target.Govno"})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	expected := []campaignbus.TargetMissingVars{
		{
			TargetID:   targetID,
			EmployeeID: employeeID,
			Vars:       []string{"Target.Govno"},
		},
	}
	if !reflect.DeepEqual(missing, expected) {
		t.Fatalf("Validate missing = %#v, want %#v", missing, expected)
	}

	if targetQuerier.linkCalls != 0 {
		t.Fatalf("target link calls = %d, want 0", targetQuerier.linkCalls)
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
	target        campaignbus.Target
	url           string
	err           error
	calls         int
	linkCalls     int
	campaignCalls int
	campaign      campaignbus.Campaign
}

func (s *targetQuerierStub) QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error) {
	s.calls++

	if s.err != nil {
		return campaignbus.Target{}, s.err
	}

	return s.target, nil
}

func (s *targetQuerierStub) PhishingURL(ctx context.Context, id uuid.UUID) (string, error) {
	s.linkCalls++

	if s.err != nil {
		return "", s.err
	}

	return s.url, nil
}

func (s *targetQuerierStub) QueryCampaignByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error) {
	s.campaignCalls++

	if s.err != nil {
		return campaignbus.Campaign{}, s.err
	}

	return s.campaign, nil
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

type employeeQuerierByIDStub struct {
	employees map[uuid.UUID]employeebus.Employee
	calls     int
}

func (s *employeeQuerierByIDStub) QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error) {
	s.calls++

	employee, ok := s.employees[id]
	if !ok {
		return employeebus.Employee{}, employeebus.ErrNotFound
	}

	return employee, nil
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

type attachmentQuerierStub struct {
	attachment attachmentbus.Attachment
	err        error
}

func (s *attachmentQuerierStub) QueryByID(_ context.Context, _ uuid.UUID) (attachmentbus.Attachment, error) {
	if s.err != nil {
		return attachmentbus.Attachment{}, s.err
	}
	return s.attachment, nil
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

