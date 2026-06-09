package campaignapi_test

import (
	"testing"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func Test_Campaign(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_Campaign")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %v", err)
	}

	// POST /campaign
	test.Run(t, create201(sd), "create-201")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// GET /campaign  GET /campaign/:id
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "querybyid-200")
	test.Run(t, queryByID404(sd), "querybyid-404")

	// PUT /campaign/:id
	test.Run(t, update200(sd), "update-200")
	test.Run(t, update401(sd), "update-401")

	// PUT /campaign/:id/start|pause|cancel
	test.Run(t, startDraft400(test.DB, sd), "start-400")
	test.Run(t, startCompleted400(test.DB, sd), "start-completed-400")
	test.Run(t, pauseDraft400(test.DB, sd), "pause-400")
	test.Run(t, cancelDraft200(test.DB, sd), "cancel-200")
	test.Run(t, cancelCanceled400(test.DB, sd), "cancel-already-canceled-400")

	// DELETE /campaign/:id
	test.Run(t, delete204(test.DB, sd), "delete-204")
	test.Run(t, delete404(sd), "delete-404")
}
