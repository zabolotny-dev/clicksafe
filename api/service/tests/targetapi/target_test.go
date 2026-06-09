package targetapi_test

import (
	"testing"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func Test_Target(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_Target")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %v", err)
	}

	// POST /target
	test.Run(t, create201(sd), "create-201")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// PUT /target/:id/schedule (must run before campaign-wide delete)
	test.Run(t, schedule200(sd), "schedule-200")
	test.Run(t, schedule401(sd), "schedule-401")

	// DELETE /target/:id
	test.Run(t, deleteByID204(test.DB, sd), "delete-byid-204")
	test.Run(t, deleteByID404(sd), "delete-byid-404")

	// DELETE /target/campaign/:campaign_id
	test.Run(t, deleteByCampaign204(test.DB, sd), "delete-bycampaign-204")
	test.Run(t, deleteByCampaign404(sd), "delete-bycampaign-404")
}
