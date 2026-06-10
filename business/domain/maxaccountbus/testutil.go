package maxaccountbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// TestSeedAccounts creates n accounts directly via the storer and returns them.
// It bypasses the adapter so it works in tests with a no-op adapter.
func TestSeedAccounts(ctx context.Context, n int, api *Business) ([]Account, error) {
	accounts := make([]Account, 0, n)

	for i := range n {
		acc := Account{
			AdapterID: uuid.New(),
			Phone:     fmt.Sprintf("+7900%07d", i),
			Label:     fmt.Sprintf("Test Account %d", i),
			Status:    StatusActive,
		}

		saved, err := api.storer.Upsert(ctx, acc)
		if err != nil {
			return nil, fmt.Errorf("seeding account idx[%d]: %w", i, err)
		}

		accounts = append(accounts, saved)
	}

	return accounts, nil
}
