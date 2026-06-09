package targetapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/targetapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/target",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &targetapp.NewTarget{
				EmployeeID: sd.EmployeeForPost.ID.String(),
				CampaignID: sd.Campaign.ID.String(),
			},
			GotResp: &targetapp.Target{},
			ExpResp: &targetapp.Target{
				EmployeeID: sd.EmployeeForPost.ID,
				CampaignID: sd.Campaign.ID,
				Status:     "PENDING",
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*targetapp.Target)
				e := exp.(*targetapp.Target)
				e.ID = g.ID
				e.Token = g.Token
				return cmp.Diff(g, e, cmpopts.IgnoreFields(targetapp.Target{}, "ScheduledAt", "CreatedAt"))
			},
		},
	}
}

func create400(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-employee-id",
			URL:        "/target",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &targetapp.NewTarget{
				EmployeeID: "not-a-uuid",
				CampaignID: sd.Campaign.ID.String(),
			},
		},
		{
			Name:       "invalid-campaign-id",
			URL:        "/target",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &targetapp.NewTarget{
				EmployeeID: sd.EmployeeForPost.ID.String(),
				CampaignID: "not-a-uuid",
			},
		},
	}
}

func create401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/target",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &targetapp.NewTarget{
				EmployeeID: sd.EmployeeForPost.ID.String(),
				CampaignID: sd.Campaign.ID.String(),
			},
		},
	}
}
