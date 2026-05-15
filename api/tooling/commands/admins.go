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
)

func Admins(cfg database.Config, login, fullname string, timeOut time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	adminStore := admindb.NewStore(db)
	adminBus := adminbus.NewBusiness(nil, adminStore)

	var filter adminbus.AdminQueryFilter
	if login != "" {
		filter.Login = &login
	}
	if fullname != "" {
		filter.FullName = &fullname
	}

	admins, err := adminBus.Query(ctx, filter)
	if err != nil {
		return fmt.Errorf("query admins: %w", err)
	}

	return json.NewEncoder(os.Stdout).Encode(admins)
}
