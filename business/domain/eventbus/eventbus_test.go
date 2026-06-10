package eventbus_test

import (
	"context"
	"fmt"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
)

func Test_Event(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Event")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, publish(db.BusDomain, sd), "publish")
	unittest.Run(t, parseEventType(), "parse")
}

// =============================================================================

type seedData struct {
	Campaign campaignbus.Campaign
	Employee employeebus.Employee
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	deps, err := departmentbus.TestSeedDepartments(ctx, 1, busDomain.Department)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding department: %w", err)
	}

	emps, err := employeebus.TestSeedEmployees(ctx, 1, &deps[0].ID, busDomain.Employee)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employee: %w", err)
	}

	camps, err := campaignbus.TestSeedCampaigns(ctx, 1, busDomain.Campaign)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding campaign: %w", err)
	}

	return seedData{Campaign: camps[0], Employee: emps[0]}, nil
}

// =============================================================================

func publish(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	ip := netip.MustParseAddr("1.2.3.4")

	return []unittest.Table{
		{
			Name:    "message-sent",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.MessageSent,
					IPAddress:  ip,
					UserAgent:  "TestAgent",
					Referer:    "https://example.com",
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "email-opened",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.EmailOpened,
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "link-opened",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.LinkOpened,
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "data-sent",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.DataSent,
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "delivery-failed",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.DeliveryFailed,
				}) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "message-read",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Event.Publish(ctx, eventbus.NewEvent{
					CampaignID: sd.Campaign.ID,
					EmployeeID: sd.Employee.ID,
					Type:       eventbus.MessageRead,
					IPAddress:  ip,
				}) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func parseEventType() []unittest.Table {
	return []unittest.Table{
		{
			Name:    "message-sent",
			ExpResp: eventbus.MessageSent.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("MESSAGE_SENT")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "email-opened",
			ExpResp: eventbus.EmailOpened.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("EMAIL_OPENED")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "link-opened",
			ExpResp: eventbus.LinkOpened.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("LINK_OPENED")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "data-sent",
			ExpResp: eventbus.DataSent.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("DATA_SENT")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "message-replied",
			ExpResp: eventbus.MessageReplied.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("MESSAGE_REPLIED")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "delivery-failed",
			ExpResp: eventbus.DeliveryFailed.String(),
			ExcFunc: func(ctx context.Context) any {
				et, err := eventbus.Parse("DELIVERY_FAILED")
				if err != nil {
					return err
				}
				return et.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "invalid-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := eventbus.Parse("INVALID_EVENT")
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
