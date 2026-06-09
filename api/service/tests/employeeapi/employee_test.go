package employeeapi_test

import (
	"testing"

	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func Test_Employee(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_Employee")

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %v", err)
	}

	// POST /employee
	test.Run(t, create201(sd), "create-201")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// GET /employee  GET /employee/:id
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "querybyid-200")
	test.Run(t, queryByID404(sd), "querybyid-404")

	// PUT /employee/:id
	test.Run(t, update200(sd), "update-200")

	// DELETE /employee/:id
	test.Run(t, delete204(test.DB, sd), "delete-204")
	test.Run(t, delete404(sd), "delete-404")
}
