package organizationapi_test

import (
	"testing"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func Test_Organization(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_Organization")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %v", err)
	}

	// POST /organization  GET /organization  PUT /organization
	test.Run(t, crud(sd), "crud")
}
