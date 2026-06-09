package organizationapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/organizationapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func crud(sd seedData) []apitest.Table {
	newLabel := "Updated Organization"

	return []apitest.Table{
		{
			Name:       "create-401",
			URL:        "/organization",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &organizationapp.NewOrganization{Label: "Unauthorized"},
		},
		{
			Name:       "create-400",
			URL:        "/organization",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &organizationapp.NewOrganization{},
		},
		{
			Name:       "create-201",
			URL:        "/organization",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &organizationapp.NewOrganization{Label: "My Organization"},
			GotResp:    &organizationapp.Organization{},
			ExpResp:    &organizationapp.Organization{Label: "My Organization"},
			CmpFunc: func(got, exp any) string {
				g := got.(*organizationapp.Organization)
				e := exp.(*organizationapp.Organization)
				e.ID = g.ID
				return cmp.Diff(g, e, cmpopts.IgnoreFields(organizationapp.Organization{},
					"AttachmentID", "Attributes"))
			},
		},
		{
			Name:       "create-409",
			URL:        "/organization",
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &organizationapp.NewOrganization{Label: "Duplicate Organization"},
		},
		{
			Name:       "get-200",
			URL:        "/organization",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &organizationapp.Organization{},
			ExpResp:    &organizationapp.Organization{Label: "My Organization"},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(organizationapp.Organization{},
					"ID", "AttachmentID", "Attributes"))
			},
		},
		{
			Name:       "get-401",
			URL:        "/organization",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
		},
		{
			Name:       "update-200",
			URL:        "/organization",
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &organizationapp.UpdateOrganization{Label: &newLabel},
			GotResp:    &organizationapp.Organization{},
			ExpResp:    &organizationapp.Organization{Label: newLabel},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(organizationapp.Organization{},
					"ID", "AttachmentID", "Attributes"))
			},
		},
		{
			Name:       "update-401",
			URL:        "/organization",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &organizationapp.UpdateOrganization{Label: &newLabel},
		},
	}
}
