package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	seedsdk "github.com/zabolotny-dev/clicksafe/business/sdk/seed"
)

type SeedConfig = seedsdk.Config

func Seed(cfg database.Config, timeOut time.Duration, seedConfig SeedConfig) error {
	seedConfig = seedsdk.DefaultConfig(seedConfig)

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := seedsdk.Run(ctx, db, seedConfig); err != nil {
		return err
	}

	fmt.Printf("%s seed complete\n", seedConfig.Scenario)
	return nil
}
