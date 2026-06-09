package campaignapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/campaignapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "email-campaign",
			URL:        "/campaign",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &campaignapp.NewCampaign{
				Label:    "Test Email Campaign",
				Type:     "EMAIL",
				Domain:   "https://phish.example.com",
				DateFrom: time.Now().Add(time.Hour),
				DateTo:   time.Now().Add(48 * time.Hour),
			},
			GotResp: &campaignapp.Campaign{},
			ExpResp: &campaignapp.Campaign{
				Label:  "Test Email Campaign",
				Type:   "EMAIL",
				Domain: "https://phish.example.com",
				Status: "DRAFT",
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*campaignapp.Campaign)
				e := exp.(*campaignapp.Campaign)
				e.ID = g.ID
				e.DateFrom = g.DateFrom
				e.DateTo = g.DateTo
				return cmp.Diff(g, e, cmpopts.IgnoreFields(campaignapp.Campaign{}, "Attributes"))
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-label",
			URL:        "/campaign",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &campaignapp.NewCampaign{},
		},
		{
			Name:       "invalid-domain",
			URL:        "/campaign",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &campaignapp.NewCampaign{
				Label:    "Bad Campaign",
				Type:     "email",
				Domain:   "not-a-url",
				DateFrom: time.Now().Add(time.Hour),
				DateTo:   time.Now().Add(48 * time.Hour),
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/campaign",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &campaignapp.NewCampaign{
				Label: "Should be rejected",
			},
		},
	}
}
