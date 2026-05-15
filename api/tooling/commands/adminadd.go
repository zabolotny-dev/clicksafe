package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus/stores/admindb"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
	passhash "github.com/zabolotny-dev/clicksafe/foundation/password"
)

func UserAdd(cfg database.Config, firstname, lastname, email, pass string, timeOut time.Duration, ph passhash.Argon2idConfig) error {
	if firstname == "" || lastname == "" || email == "" || pass == "" {
		fmt.Println("help: adminadd <firstname> <lastname> <login> <password>")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	hasher, err := passhash.New(ph)
	if err != nil {
		return fmt.Errorf("new hasher: %w", err)
	}

	adminStore := admindb.NewStore(db)
	adminBus := adminbus.NewBusiness(hasher, adminStore)

	fname, err := name.Parse(firstname)
	if err != nil {
		return fmt.Errorf("parse first name: %w", err)
	}
	lname, err := name.Parse(lastname)
	if err != nil {
		return fmt.Errorf("parse last name: %w", err)
	}
	log, err := login.Parse(email)
	if err != nil {
		return fmt.Errorf("parse login: %w", err)
	}
	ps, err := password.Parse(pass)
	if err != nil {
		return fmt.Errorf("parse password: %w", err)
	}

	admin, err := adminBus.Save(ctx, adminbus.NewAdmin{
		FirstName: fname,
		LastName:  lname,
		Login:     log,
		Password:  ps,
	})
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(admin)
}
