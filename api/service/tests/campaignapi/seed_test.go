package campaignapi_test

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

func insertSeedData(db *dbtest.Database) (apitest.SeedData, error) {
	ctx := context.Background()

	adm, err := db.BusDomain.Admin.Save(ctx, adminbus.NewAdmin{
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("Admin"),
		Login:     login.MustParse("test.admin"),
		Password:  password.MustParse("TestPass123456!"),
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating admin: %w", err)
	}

	sess, err := db.BusDomain.Session.Create(ctx, sessionbus.NewSession{
		AdminID:   adm.ID,
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "apitest",
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("creating session: %w", err)
	}

	cmp, err := db.BusDomain.Campaign.Save(ctx, campaignbus.NewCampaign{
		Label: label.MustParse("Seeded Campaign"),
	})
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding campaign: %w", err)
	}

	return apitest.SeedData{
		Admins: []apitest.SeedAdmin{
			{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			},
		},
		Campaigns: []campaignbus.Campaign{cmp},
	}, nil
}
