package landingapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/domain/landingapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func query200(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/landing",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func queryByID200(sd seedData) []apitest.Table {
	landing := sd.Landings[0]

	return []apitest.Table{
		{
			Name:       "seeded",
			URL:        fmt.Sprintf("/landing/%s", landing.ID),
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &landingapp.Landing{},
			ExpResp:    &landingapp.Landing{ID: landing.ID, Label: landing.Label.String()},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(landingapp.Landing{}, "HtmlBodyID"))
			},
		},
	}
}

func queryByID404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/landing/%s", uuid.New()),
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
