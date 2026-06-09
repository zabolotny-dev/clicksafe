package employeeapi_test

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

type seedData struct {
	apitest.SeedData
	Employees []employeebus.Employee
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

	emp, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName: name.MustParse("John"),
		LastName:  name.MustParse("Doe"),
		Email:     mail.Address{Address: "john.doe@example.com"},
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employee: %w", err)
	}

	return seedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.SeedAdmin{{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			}},
		},
		Employees: []employeebus.Employee{emp},
	}, nil
}
