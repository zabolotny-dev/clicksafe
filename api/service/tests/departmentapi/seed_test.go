package departmentapi_test

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

type seedData struct {
	apitest.SeedData
	Departments []departmentbus.Department
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

	dep, err := db.BusDomain.Department.Save(ctx, departmentbus.NewDepartment{
		Label: label.MustParse("Seeded Department"),
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding department: %w", err)
	}

	return seedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.SeedAdmin{{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			}},
		},
		Departments: []departmentbus.Department{dep},
	}, nil
}
