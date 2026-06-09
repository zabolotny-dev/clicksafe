package campaignapi_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func delete204(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	// Create a fresh campaign to delete so the seeded one remains intact for other tests.
	cmp, err := db.BusDomain.Campaign.Save(context.Background(), campaignbus.NewCampaign{
		Label: label.MustParse("To Be Deleted"),
	})
	if err != nil {
		panic("delete_test seed: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "existing",
			URL:        fmt.Sprintf("/campaign/%s", cmp.ID),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func delete404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/campaign/%s", uuid.New()),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
