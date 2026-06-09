package messageapi_test

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

type seedData struct {
	apitest.SeedData
	Messages []messagebus.Message
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

	fromName, err := label.ParseNull("NoReply")
	if err != nil {
		return seedData{}, fmt.Errorf("parsing from_name: %w", err)
	}

	sub, err := subject.ParseNull("Test Subject")
	if err != nil {
		return seedData{}, fmt.Errorf("parsing subject: %w", err)
	}

	msg, err := db.BusDomain.Message.Save(ctx, messagebus.NewMessage{
		Label:     label.MustParse("Seed Message"),
		FromEmail: mail.Address{Address: "noreply@example.com"},
		FromName:  fromName,
		Subject:   sub,
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding message: %w", err)
	}

	return seedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.SeedAdmin{{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			}},
		},
		Messages: []messagebus.Message{msg},
	}, nil
}
