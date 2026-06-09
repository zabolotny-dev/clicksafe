package targetapi_test

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"
	"time"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

type seedData struct {
	apitest.SeedData
	Campaign        campaignbus.Campaign
	Employee        employeebus.Employee
	EmployeeForPost employeebus.Employee
	Target          campaignbus.Target
}

func insertSeedData(db *dbtest.Database) (seedData, error) {
	ctx := context.Background()

	adm, err := db.BusDomain.Admin.Save(ctx, adminbus.NewAdmin{
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("Admin"),
		Login:     login.MustParse("test.admin"),
		Password:  password.MustParse("TestPass123456!"),
	})
	if err != nil {
		return seedData{}, fmt.Errorf("creating admin: %w", err)
	}

	sess, err := db.BusDomain.Session.Create(ctx, sessionbus.NewSession{
		AdminID:   adm.ID,
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "apitest",
	})
	if err != nil {
		return seedData{}, fmt.Errorf("creating session: %w", err)
	}

	now := time.Now().UTC()
	dateRange, err := date.ParseNull(now.Add(-time.Hour), now.Add(30*24*time.Hour))
	if err != nil {
		return seedData{}, fmt.Errorf("parsing date range: %w", err)
	}

	campaign, err := db.BusDomain.Campaign.Save(ctx, campaignbus.NewCampaign{
		Label:     label.MustParse("Seed Campaign"),
		DateRange: dateRange,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding campaign: %w", err)
	}

	employee, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName: name.MustParse("John"),
		LastName:  name.MustParse("Doe"),
		Email:     mail.Address{Address: "john.doe@example.com"},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employee: %w", err)
	}

	employeeForPost, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName: name.MustParse("Jane"),
		LastName:  name.MustParse("Smith"),
		Email:     mail.Address{Address: "jane.smith@example.com"},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employee for post: %w", err)
	}

	target, err := db.BusDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: campaign.ID,
		EmployeeID: employee.ID,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding target: %w", err)
	}

	return seedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.SeedAdmin{{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			}},
		},
		Campaign:        campaign,
		Employee:        employee,
		EmployeeForPost: employeeForPost,
		Target:          target,
	}, nil
}
