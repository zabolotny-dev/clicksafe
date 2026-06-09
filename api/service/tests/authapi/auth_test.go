package authapi_test

import (
	"testing"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func Test_Auth(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_Auth")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %v", err)
	}

	// POST /login
	test.Run(t, login200(sd), "login-200")
	test.Run(t, login401(sd), "login-401")

	// GET /me
	test.Run(t, me200(sd), "me-200")
	test.Run(t, me401(sd), "me-401")

	// POST /logout
	test.Run(t, logout200(sd), "logout-200")
}
