package authapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/authapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func me200(sd seedData) []apitest.Table {
	adm := sd.Admins[0]

	return []apitest.Table{
		{
			Name:       "with-session",
			URL:        "/me",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      adm.RawToken,
			CSRFToken:  adm.CSRFToken,
			GotResp:    &authapp.Me{},
			ExpResp: &authapp.Me{
				FirstName: adm.Admin.FirstName.String(),
				LastName:  adm.Admin.LastName.String(),
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(authapp.Me{}, "CSRFToken", "ExpiresAt"))
			},
		},
	}
}

func me401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-session",
			URL:        "/me",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
		},
	}
}

func logout200(sd seedData) []apitest.Table {
	adm := sd.Admins[0]

	return []apitest.Table{
		{
			Name:       "with-session",
			URL:        "/logout",
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
			Token:      adm.RawToken,
			CSRFToken:  adm.CSRFToken,
		},
	}
}
