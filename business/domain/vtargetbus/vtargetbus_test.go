package vtargetbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
)

// vtargetStorerStub implements Storer for error-path unit tests.
type vtargetStorerStub struct {
	queryErr error
	countErr error
}

func (s *vtargetStorerStub) Query(_ context.Context, _ vtargetbus.Filter, _ order.By, _ page.Page) ([]vtargetbus.Target, error) {
	return nil, s.queryErr
}

func (s *vtargetStorerStub) Count(_ context.Context, _ vtargetbus.Filter) (int, error) {
	return 0, s.countErr
}

func Test_VTarget_ErrorPaths(t *testing.T) {
	t.Parallel()
	unittest.Run(t, vtargetErrorPaths(), "error-paths")
}

func vtargetErrorPaths() []unittest.Table {
	dbErr := errors.New("db error")

	busQueryErr := vtargetbus.NewBusiness(&vtargetStorerStub{queryErr: dbErr})
	busCountErr := vtargetbus.NewBusiness(&vtargetStorerStub{countErr: dbErr})

	return []unittest.Table{
		{
			Name:    "query-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQueryErr.Query(ctx, vtargetbus.Filter{}, vtargetbus.DefaultOrderBy, page.MustParse("1", "10"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "count-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busCountErr.Count(ctx, vtargetbus.Filter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func Test_VTarget(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_VTarget")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, query(db.BusDomain, sd), "query")
}

// =============================================================================

type seedData struct {
	Campaign  campaignbus.Campaign
	Employees []employeebus.Employee
	Targets   []campaignbus.Target
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	deps, err := departmentbus.TestSeedDepartments(ctx, 1, busDomain.Department)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding department: %w", err)
	}

	emps, err := employeebus.TestSeedEmployees(ctx, 2, &deps[0].ID, busDomain.Employee)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employees: %w", err)
	}

	camps, err := campaignbus.TestSeedCampaigns(ctx, 1, busDomain.Campaign)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding campaign: %w", err)
	}
	camp := camps[0]

	targets := make([]campaignbus.Target, 0, len(emps))
	for _, emp := range emps {
		t2, err := busDomain.Target.Save(ctx, campaignbus.NewTarget{
			CampaignID: camp.ID,
			EmployeeID: emp.ID,
		})
		if err != nil {
			return seedData{}, fmt.Errorf("seeding target for employee %s: %w", emp.ID, err)
		}
		targets = append(targets, t2)
	}

	return seedData{Campaign: camp, Employees: emps, Targets: targets}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "bycampaign-returns-targets",
			ExpResp: len(sd.Targets),
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.VTarget.Query(ctx,
					vtargetbus.Filter{CampaignID: &sd.Campaign.ID},
					vtargetbus.DefaultOrderBy,
					page.MustParse("1", "10"),
				)
				if err != nil {
					return err
				}
				return len(resp)
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "count-bycampaign",
			ExpResp: len(sd.Targets),
			ExcFunc: func(ctx context.Context) any {
				count, err := busDomain.VTarget.Count(ctx, vtargetbus.Filter{CampaignID: &sd.Campaign.ID})
				if err != nil {
					return err
				}
				return count
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
