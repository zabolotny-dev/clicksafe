package campaignbus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// TestSeedCampaigns creates n draft email campaigns via the business layer.
func TestSeedCampaigns(ctx context.Context, n int, api *CampaignBusiness) ([]Campaign, error) {
	camps := make([]Campaign, 0, n)

	idx := rand.Intn(10000)
	for range n {
		idx++

		nc := NewCampaign{
			Type:       EmailCampaign,
			Label:      label.MustParse(fmt.Sprintf("Campaign %d", idx)),
			Attributes: map[string]string{},
		}

		camp, err := api.Save(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding campaign idx[%d]: %w", idx, err)
		}

		camps = append(camps, camp)
	}

	return camps, nil
}
