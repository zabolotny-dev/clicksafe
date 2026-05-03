package resolverbus

import (
	"context"
	"errors"
	"net/mail"
	"reflect"
	"testing"

	"github.com/google/uuid"
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
	employeeBus := &employeeQuerierStub{employee: testEmployee(&departmentID)}
	departmentBus := &departmentQuerierStub{department: testDepartment(departmentID)}
	organizationBus := &organizationGetterStub{organization: testOrganization()}

	business := NewBusiness(employeeBus, departmentBus, organizationBus)

	result, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{
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

	employeeBus := &employeeQuerierStub{employee: testEmployee(nil)}
	departmentBus := &departmentQuerierStub{}
	organizationBus := &organizationGetterStub{}

	business := NewBusiness(employeeBus, departmentBus, organizationBus)

	result, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{
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

	if departmentBus.calls != 0 {
		t.Fatalf("department bus calls = %d, want 0", departmentBus.calls)
	}

	if organizationBus.calls != 0 {
		t.Fatalf("organization bus calls = %d, want 0", organizationBus.calls)
	}
}

func TestResolveEmptyPathsDoesNotRequireEmployee(t *testing.T) {
	t.Parallel()

	employeeBus := &employeeQuerierStub{err: errors.New("should not be called")}
	business := NewBusiness(employeeBus, &departmentQuerierStub{}, &organizationGetterStub{})

	result, err := business.Resolve(context.Background(), Scope{}, nil)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	expected := Result{Data: map[string]any{}}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Resolve result = %#v, want %#v", result, expected)
	}

	if employeeBus.calls != 0 {
		t.Fatalf("employee bus calls = %d, want 0", employeeBus.calls)
	}
}

func TestResolveRequiresEmployeeID(t *testing.T) {
	t.Parallel()

	business := NewBusiness(&employeeQuerierStub{}, &departmentQuerierStub{}, &organizationGetterStub{})

	_, err := business.Resolve(context.Background(), Scope{}, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrEmployeeIDRequired) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrEmployeeIDRequired)
	}
}

func TestResolveMapsEmployeeNotFound(t *testing.T) {
	t.Parallel()

	business := NewBusiness(
		&employeeQuerierStub{err: employeebus.ErrNotFound},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{"Employee.FirstName"})
	if !errors.Is(err, ErrEmployeeNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrEmployeeNotFound)
	}
}

func TestResolveRejectsInvalidPaths(t *testing.T) {
	t.Parallel()

	paths := []string{"", " ", "Employee", "Employee.", "Employee..FirstName", ".Employee.FirstName"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			business := NewBusiness(
				&employeeQuerierStub{employee: testEmployee(nil)},
				&departmentQuerierStub{},
				&organizationGetterStub{},
			)

			_, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{path})
			if !errors.Is(err, ErrUnsupportedPath) {
				t.Fatalf("Resolve error = %v, want %v", err, ErrUnsupportedPath)
			}
		})
	}
}

func TestResolveRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	business := NewBusiness(
		&employeeQuerierStub{employee: testEmployee(nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{"Campaign.Label"})
	if !errors.Is(err, ErrUnsupportedPath) {
		t.Fatalf("Resolve error = %v, want %v", err, ErrUnsupportedPath)
	}
}

func TestResolveReturnsDepartmentLookupError(t *testing.T) {
	t.Parallel()

	departmentID := uuid.New()
	business := NewBusiness(
		&employeeQuerierStub{employee: testEmployee(&departmentID)},
		&departmentQuerierStub{err: departmentbus.ErrNotFound},
		&organizationGetterStub{},
	)

	_, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{"Department.Label"})
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

	business := NewBusiness(
		&employeeQuerierStub{employee: testEmployee(nil)},
		&departmentQuerierStub{},
		&organizationGetterStub{err: organizationbus.ErrNotFound},
	)

	_, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{"Organization.Label"})
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

	employee := testEmployee(nil)
	employee.Phone = phone.Null{}
	business := NewBusiness(
		&employeeQuerierStub{employee: employee},
		&departmentQuerierStub{},
		&organizationGetterStub{},
	)

	result, err := business.Resolve(context.Background(), Scope{EmployeeID: uuid.New()}, []string{"Employee.Phone"})
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

func testEmployee(departmentID *uuid.UUID) employeebus.Employee {
	phoneNumber, err := phone.ParseNull("+79991234567")
	if err != nil {
		panic(err)
	}

	return employeebus.Employee{
		ID:           uuid.New(),
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
