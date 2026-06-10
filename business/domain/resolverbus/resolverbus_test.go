package resolverbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

func Test_Resolver(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Resolver")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, resolve(db.BusDomain, sd), "resolve")
	unittest.Run(t, validate(db.BusDomain, sd), "validate")
}

// =============================================================================

type seedData struct {
	// employee with department, campaign without domain
	Target     campaignbus.Target
	Employee   employeebus.Employee
	Department departmentbus.Department
	Campaign   campaignbus.Campaign

	// employee without department (for missing-department tests)
	TargetNoDept  campaignbus.Target
	EmployeeNoDept employeebus.Employee

	// campaign with domain (for Target.Link tests)
	CampaignWithDomain campaignbus.Campaign
	TargetWithDomain   campaignbus.Target

	// second employee in same department (for caching tests)
	Employee2       employeebus.Employee
	TargetEmployee2 campaignbus.Target
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	_, err := busDomain.Organization.Save(ctx, organizationbus.NewOrganization{
		Label:      label.MustParse("Test Organization"),
		Attributes: map[string]string{"Region": "Moscow"},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding organization: %w", err)
	}

	deps, err := departmentbus.TestSeedDepartments(ctx, 1, busDomain.Department)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding department: %w", err)
	}
	dep := deps[0]

	// two employees with department
	empsWithDept, err := employeebus.TestSeedEmployees(ctx, 2, &dep.ID, busDomain.Employee)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employees with dept: %w", err)
	}

	// one employee without department (manual seed to avoid email conflicts)
	ph, _ := phone.ParseNull("")
	empNoDept, err := busDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName:  name.MustParse("NoDept"),
		LastName:   name.MustParse("Employee"),
		Email:      mail.Address{Address: "nodept@resolver-test.example.com"},
		Phone:      ph,
		Attributes: map[string]string{},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employee without dept: %w", err)
	}
	empsNoDept := []employeebus.Employee{empNoDept}

	camps, err := busDomain.Campaign.Save(ctx, campaignbus.NewCampaign{
		Type:       campaignbus.EmailCampaign,
		Label:      label.MustParse("Campaign No Domain"),
		Attributes: map[string]string{},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding campaign: %w", err)
	}

	campWithDomain, err := busDomain.Campaign.Save(ctx, campaignbus.NewCampaign{
		Type:       campaignbus.EmailCampaign,
		Label:      label.MustParse("Campaign With Domain"),
		Domain:     domain.MustParse("https://phish.example.com"),
		Attributes: map[string]string{},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding campaign with domain: %w", err)
	}

	target, err := busDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: camps.ID,
		EmployeeID: empsWithDept[0].ID,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding target: %w", err)
	}

	targetNoDept, err := busDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: camps.ID,
		EmployeeID: empsNoDept[0].ID,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding target no dept: %w", err)
	}

	targetWithDomain, err := busDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: campWithDomain.ID,
		EmployeeID: empsWithDept[0].ID,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding target with domain: %w", err)
	}

	targetEmployee2, err := busDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: camps.ID,
		EmployeeID: empsWithDept[1].ID,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding target employee2: %w", err)
	}

	return seedData{
		Target:             target,
		Employee:           empsWithDept[0],
		Department:         dep,
		Campaign:           camps,
		TargetNoDept:       targetNoDept,
		EmployeeNoDept:     empsNoDept[0],
		CampaignWithDomain: campWithDomain,
		TargetWithDomain:   targetWithDomain,
		Employee2:          empsWithDept[1],
		TargetEmployee2:    targetEmployee2,
	}, nil
}

// nestedGet navigates a nested map[string]any by dot-separated path.
func nestedGet(data map[string]any, keys ...string) (any, bool) {
	var cur any = data
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[k]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// =============================================================================

func resolve(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "empty-paths-returns-nil",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, nil)
				return err == nil && data == nil && missing == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "nil-target-id-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, _, err := busDomain.Resolver.Resolve(ctx, uuid.Nil, []string{"Employee.FirstName"})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-target-returns-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, _, err := busDomain.Resolver.Resolve(ctx, uuid.New(), []string{"Employee.FirstName"})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-firstname",
			ExpResp: sd.Employee.FirstName.String(),
			ExcFunc: func(ctx context.Context) any {
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Employee.FirstName"})
				if err != nil {
					return err
				}
				if len(missing) > 0 {
					return fmt.Sprintf("missing: %v", missing)
				}
				val, ok := nestedGet(data, "Employee", "FirstName")
				if !ok {
					return "Employee.FirstName not found in nested data"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-lastname",
			ExpResp: sd.Employee.LastName.String(),
			ExcFunc: func(ctx context.Context) any {
				data, _, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Employee.LastName"})
				if err != nil {
					return err
				}
				val, ok := nestedGet(data, "Employee", "LastName")
				if !ok {
					return "Employee.LastName not found"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-email",
			ExpResp: sd.Employee.Email.Address,
			ExcFunc: func(ctx context.Context) any {
				data, _, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Employee.Email.Address"})
				if err != nil {
					return err
				}
				val, ok := nestedGet(data, "Employee", "Email", "Address")
				if !ok {
					return "Employee.Email.Address not found"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-phone-missing-when-empty",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// seeded employees have empty (invalid) phone → isMissingOptionalValue
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Employee.Phone"})
				return err == nil && len(missing) == 1 && missing[0] == "Employee.Phone"
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "department-label",
			ExpResp: sd.Department.Label.String(),
			ExcFunc: func(ctx context.Context) any {
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Department.Label"})
				if err != nil {
					return err
				}
				if len(missing) > 0 {
					return fmt.Sprintf("missing: %v", missing)
				}
				val, ok := nestedGet(data, "Department", "Label")
				if !ok {
					return "Department.Label not found"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "department-missing-when-employee-has-no-dept",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetNoDept.ID, []string{"Department.Label"})
				return err == nil && len(missing) == 1 && missing[0] == "Department.Label"
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "department-cached-on-second-path",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// resolving two Department paths uses the same loadDepartment call (cached)
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Department.Label", "Department.Label"})
				return err == nil && len(missing) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "organization-label",
			ExpResp: "Test Organization",
			ExcFunc: func(ctx context.Context) any {
				data, _, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Organization.Label"})
				if err != nil {
					return err
				}
				val, ok := nestedGet(data, "Organization", "Label")
				if !ok {
					return "Organization.Label not found"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "organization-attributes-map-walk",
			ExpResp: "Moscow",
			ExcFunc: func(ctx context.Context) any {
				// Organization.Attributes.Region tests the map-walk path in walkSegment
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Organization.Attributes.Region"})
				if err != nil {
					return err
				}
				if len(missing) > 0 {
					return fmt.Sprintf("missing: %v", missing)
				}
				val, ok := nestedGet(data, "Organization", "Attributes", "Region")
				if !ok {
					return "Organization.Attributes.Region not found"
				}
				return fmt.Sprintf("%v", val)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "organization-attributes-missing-key",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// map lookup for a non-existent key → walkSegment map case returns missing
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Organization.Attributes.NonExistentKey"})
				return err == nil && len(missing) == 1
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-link",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Target.Link requires campaign with domain; loadTarget is exercised
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetWithDomain.ID,
					[]string{"Target.Link"})
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				if len(missing) > 0 {
					return fmt.Sprintf("missing: %v", missing)
				}
				val, ok := nestedGet(data, "Target", "Link")
				return ok && fmt.Sprintf("%v", val) != ""
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-non-link-field-is-missing",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// rootTarget with non-Link segment → returns empty targetData → field missing
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Target.Status"})
				return err == nil && len(missing) == 1
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-domain",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// loadCampaign is exercised; Domain with value → non-missing
				data, _, err := busDomain.Resolver.Resolve(ctx, sd.TargetWithDomain.ID, []string{"Campaign.Domain"})
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				val, ok := nestedGet(data, "Campaign", "Domain")
				return ok && fmt.Sprintf("%v", val) != ""
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-link-no-domain-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// sd.Target's campaign has no domain → PhishingURL returns ErrDomainRequired
				// loadTarget catches it and wraps as ErrDomainRequired
				_, _, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Target.Link"})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-root-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, _, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID, []string{"Unknown.Field"})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-path-in-missing",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Employee.NonExistentField"})
				return err == nil && len(missing) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "duplicate-missing-paths-deduplicated",
			ExpResp: 1,
			ExcFunc: func(ctx context.Context) any {
				// same missing path twice → only reported once
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Employee.NonExistent", "Employee.NonExistent"})
				if err != nil {
					return err
				}
				return len(missing)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "organization-attributes-json-marshal",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// resolving the map itself (2 segments) hits stringifyValue JSON marshal branch
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Organization.Attributes"})
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				if len(missing) > 0 {
					return fmt.Sprintf("missing: %v", missing)
				}
				val, ok := nestedGet(data, "Organization", "Attributes")
				// result is a JSON string like {"Region":"Moscow"}
				return ok && val != ""
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-department-id-nil-pointer-missing",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// no-dept employee has DepartmentID=nil (*uuid.UUID nil pointer)
				// dereferenceValue hits the nil pointer → isMissing=true path
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetNoDept.ID,
					[]string{"Employee.DepartmentID"})
				return err == nil && len(missing) == 1
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "department-cached-no-dept-second-path",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// depLoaded=true, hasDepartment=false on second call (no-dept employee)
				// duplicate paths are deduplicated → len(missing)==1
				_, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetNoDept.ID,
					[]string{"Department.Label", "Department.Label"})
				return err == nil && len(missing) == 1
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-link-cached-on-second-path",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// linkLoaded=true on second resolution of Target.Link
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetWithDomain.ID,
					[]string{"Target.Link", "Target.Link"})
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				// duplicated path → only inserted once, but both resolve successfully
				_, hasLink := nestedGet(data, "Target", "Link")
				return hasLink && len(missing) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-cached-on-second-path",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// campLoaded=true on second Campaign.* resolution
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.TargetWithDomain.ID,
					[]string{"Campaign.Domain", "Campaign.Domain"})
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				_, hasDomain := nestedGet(data, "Campaign", "Domain")
				return hasDomain && len(missing) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "multiple-paths-resolved-together",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				data, missing, err := busDomain.Resolver.Resolve(ctx, sd.Target.ID,
					[]string{"Employee.FirstName", "Employee.LastName", "Organization.Label"})
				if err != nil || len(missing) > 0 {
					return false
				}
				_, hasFirst := nestedGet(data, "Employee", "FirstName")
				_, hasLast := nestedGet(data, "Employee", "LastName")
				_, hasOrg := nestedGet(data, "Organization", "Label")
				return hasFirst && hasLast && hasOrg
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

// =============================================================================

func validate(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	phishDomain := domain.MustParse("https://phish.example.com")
	campWithDomain := sd.CampaignWithDomain
	_ = phishDomain

	return []unittest.Table{
		{
			Name:    "empty-targets-returns-nil",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					nil,
					[]string{"Employee.FirstName"},
				)
				return err == nil && result == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "empty-required-vars-returns-nil",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target},
					nil,
				)
				return err == nil && result == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "invalid-path-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target},
					[]string{"EmployeeOnlyNoSeparator"},
				)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-link-without-domain-returns-domain-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// needsTargetLink=true but campaign has no domain → ErrDomainRequired
				_, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign, // no domain
					[]campaignbus.Target{sd.Target},
					[]string{"Target.Link"},
				)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "all-paths-present-returns-empty-result",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target},
					[]string{"Employee.FirstName", "Department.Label"},
				)
				if err != nil {
					return err
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-link-with-domain-no-missing",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				// Target.Link with domain set → link is injected without calling PhishingURL
				result, err := busDomain.Resolver.Validate(ctx,
					campWithDomain,
					[]campaignbus.Target{sd.TargetWithDomain},
					[]string{"Target.Link"},
				)
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "missing-department-reported-for-target",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.TargetNoDept},
					[]string{"Department.Label"},
				)
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				return len(result) == 1 && result[0].TargetID == sd.TargetNoDept.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "mixed-targets-some-missing-some-present",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// targetWithDept has department, targetNoDept doesn't
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target, sd.TargetNoDept},
					[]string{"Department.Label"},
				)
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				// only one target should have missing vars
				return len(result) == 1 && result[0].TargetID == sd.TargetNoDept.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cached-department-across-multiple-targets",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				// two targets with employees in same department
				// cachedDepartmentQuerier caches dept after first query
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target, sd.TargetEmployee2},
					[]string{"Department.Label"},
				)
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cached-organization-across-multiple-targets",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				// cachedOrganizationGetter caches org after first query
				result, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target, sd.TargetEmployee2},
					[]string{"Organization.Label"},
				)
				if err != nil {
					return fmt.Sprintf("error: %v", err)
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-root-in-validate-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Resolver.Validate(ctx,
					sd.Campaign,
					[]campaignbus.Target{sd.Target},
					[]string{"Unknown.Field"},
				)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
