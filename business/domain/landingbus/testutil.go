package landingbus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// TestSeedLandings creates n landings via the business layer (without HTML body).
func TestSeedLandings(ctx context.Context, n int, api *Business) ([]Landing, error) {
	landings := make([]Landing, 0, n)

	idx := rand.Intn(10000)
	for range n {
		idx++

		nl := NewLanding{
			Label: label.MustParse(fmt.Sprintf("Landing %d", idx)),
		}

		landing, err := api.Save(ctx, nl)
		if err != nil {
			return nil, fmt.Errorf("seeding landing idx[%d]: %w", idx, err)
		}

		landings = append(landings, landing)
	}

	return landings, nil
}
