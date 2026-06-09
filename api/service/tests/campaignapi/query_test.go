package campaignapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/domain/campaignapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/campaign",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	seeded := sd.Campaigns[0]

	return []apitest.Table{
		{
			Name:       "seeded-campaign",
			URL:        fmt.Sprintf("/campaign/%s", seeded.ID),
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &campaignapp.Campaign{},
			ExpResp: &campaignapp.Campaign{
				ID:     seeded.ID,
				Label:  seeded.Label.String(),
				Type:   seeded.Type.String(),
				Status: seeded.Status.String(),
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*campaignapp.Campaign)
				e := exp.(*campaignapp.Campaign)
				return cmp.Diff(g, e,
					cmpopts.IgnoreFields(campaignapp.Campaign{},
						"Domain", "DateFrom", "DateTo", "Attributes",
						"MessageID", "LandingID", "EducationID", "MaxEducationTextID",
					),
				)
			},
		},
	}
}

func queryByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/campaign/%s", uuid.New()),
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
