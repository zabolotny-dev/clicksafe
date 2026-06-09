package campaignapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/campaignapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	seeded := sd.Campaigns[0]
	newLabel := "Updated Label"

	return []apitest.Table{
		{
			Name:       "change-label",
			URL:        fmt.Sprintf("/campaign/%s", seeded.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &campaignapp.UpdateCampaign{Label: &newLabel},
			GotResp:    &campaignapp.Campaign{},
			ExpResp: &campaignapp.Campaign{
				ID:     seeded.ID,
				Label:  newLabel,
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

func update401(sd apitest.SeedData) []apitest.Table {
	seeded := sd.Campaigns[0]
	newLabel := "Should Fail"

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/campaign/%s", seeded.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &campaignapp.UpdateCampaign{Label: &newLabel},
		},
	}
}
