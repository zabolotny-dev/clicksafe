package adminbus

import (
	"context"
	"fmt"

	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

var seedFirstNames = []string{
	"Alpha", "Beta", "Gamma", "Delta", "Epsilon",
	"Zeta", "Eta", "Theta", "Iota", "Kappa",
}

var seedLastNames = []string{
	"Smith", "Jones", "Brown", "Davis", "Wilson",
	"Taylor", "Clark", "Lewis", "Young", "Walker",
}

// TestSeedAdmins creates n admins and inserts them via the business layer.
// It also returns the plain-text passwords for each admin.
func TestSeedAdmins(ctx context.Context, n int, api *Business) ([]Admin, []string, error) {
	admins := make([]Admin, 0, n)
	passwords := make([]string, 0, n)

	for i := range n {
		idx := i % len(seedFirstNames)
		plainPwd := fmt.Sprintf("Password%dSecure!", i)

		na := NewAdmin{
			FirstName: name.MustParse(seedFirstNames[idx]),
			LastName:  name.MustParse(seedLastNames[idx]),
			Login:     login.MustParse(fmt.Sprintf("seed.admin.%03d", i)),
			Password:  password.MustParse(plainPwd),
		}

		adm, err := api.Save(ctx, na)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding admin idx[%d]: %w", i, err)
		}

		admins = append(admins, adm)
		passwords = append(passwords, plainPwd)
	}

	return admins, passwords, nil
}
