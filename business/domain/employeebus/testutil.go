package employeebus

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

var firstNames = []string{
	"Alpha", "Beta", "Gamma", "Delta", "Epsilon",
	"Zeta", "Eta", "Theta", "Iota", "Kappa",
}

var lastNames = []string{
	"Smith", "Jones", "Brown", "Davis", "Wilson",
	"Taylor", "Clark", "Lewis", "Young", "Walker",
}

// TestSeedEmployees creates n employees and inserts them via the business layer.
// deptID is optional — pass nil to create employees without a department.
func TestSeedEmployees(ctx context.Context, n int, deptID *uuid.UUID, api *Business) ([]Employee, error) {
	emps := make([]Employee, 0, n)

	for i := range n {
		idx := i % len(firstNames)
		lastName := lastNames[idx]

		ph, _ := phone.ParseNull("")

		ne := NewEmployee{
			DepartmentID: deptID,
			FirstName:    name.MustParse(firstNames[idx]),
			LastName:     name.MustParse(lastName),
			Email:        mail.Address{Address: fmt.Sprintf("emp%d@example.com", i)},
			Phone:        ph,
			Attributes:   map[string]string{},
		}

		emp, err := api.Save(ctx, ne)
		if err != nil {
			return nil, fmt.Errorf("seeding employee idx[%d]: %w", i, err)
		}

		emps = append(emps, emp)
	}

	return emps, nil
}
