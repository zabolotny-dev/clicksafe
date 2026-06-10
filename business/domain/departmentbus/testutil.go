package departmentbus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// TestSeedDepartments creates n departments and inserts them via the business layer.
func TestSeedDepartments(ctx context.Context, n int, api *Business) ([]Department, error) {
	deps := make([]Department, 0, n)

	idx := rand.Intn(10000)
	for range n {
		idx++

		nd := NewDepartment{
			Label:      label.MustParse(fmt.Sprintf("Department %d", idx)),
			Attributes: map[string]string{},
		}

		dep, err := api.Save(ctx, nd)
		if err != nil {
			return nil, fmt.Errorf("seeding department idx[%d]: %w", idx, err)
		}

		deps = append(deps, dep)
	}

	return deps, nil
}
