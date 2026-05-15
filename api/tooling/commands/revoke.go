package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus/stores/sessiondb"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
)

func RevokeSessions(cfg database.Config, adminIDStr string, timeOut time.Duration) error {
	if adminIDStr == "" {
		fmt.Println("help: revoke <admin_id>")
		return ErrHelp
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return fmt.Errorf("parse admin id: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	sessionStore := sessiondb.NewStore(db)
	sessionBus := sessionbus.NewBusiness(sessionStore, nil, 0)

	if err := sessionBus.RevokeByAdminID(ctx, adminID); err != nil {
		return fmt.Errorf("revoke by admin id: %w", err)
	}

	fmt.Printf("Successfully revoked all sessions for admin %s\n", adminIDStr)
	return nil
}
