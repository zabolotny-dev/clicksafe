package landingapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/landingapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func update200(sd seedData) []apitest.Table {
	landing := sd.Landings[0]
	newLabel := "Updated Landing"

	return []apitest.Table{
		{
			Name:       "change-label",
			URL:        fmt.Sprintf("/landing/%s", landing.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &landingapp.UpdateLanding{Label: &newLabel},
			GotResp:    &landingapp.Landing{},
			ExpResp:    &landingapp.Landing{ID: landing.ID, Label: newLabel},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(landingapp.Landing{}, "HtmlBodyID"))
			},
		},
	}
}

func update401(sd seedData) []apitest.Table {
	landing := sd.Landings[0]
	newLabel := "Should Fail"

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/landing/%s", landing.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &landingapp.UpdateLanding{Label: &newLabel},
		},
	}
}
