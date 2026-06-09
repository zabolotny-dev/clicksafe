package landingapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/landingapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/landing",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &landingapp.NewLanding{Label: "New Landing"},
			GotResp:    &landingapp.Landing{},
			ExpResp:    &landingapp.Landing{Label: "New Landing"},
			CmpFunc: func(got, exp any) string {
				g := got.(*landingapp.Landing)
				e := exp.(*landingapp.Landing)
				e.ID = g.ID
				return cmp.Diff(g, e, cmpopts.IgnoreFields(landingapp.Landing{}, "HtmlBodyID"))
			},
		},
	}
}

func create400(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-label",
			URL:        "/landing",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &landingapp.NewLanding{},
		},
	}
}

func create401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/landing",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &landingapp.NewLanding{Label: "Unauthorized"},
		},
	}
}
