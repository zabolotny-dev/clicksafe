package departmentapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/departmentapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/department",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &departmentapp.NewDepartment{Label: "New Department"},
			GotResp:    &departmentapp.Department{},
			ExpResp:    &departmentapp.Department{Label: "New Department"},
			CmpFunc: func(got, exp any) string {
				g := got.(*departmentapp.Department)
				e := exp.(*departmentapp.Department)
				e.ID = g.ID
				return cmp.Diff(g, e, cmpopts.IgnoreFields(departmentapp.Department{}, "Attributes"))
			},
		},
	}
}

func create400(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-label",
			URL:        "/department",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &departmentapp.NewDepartment{},
		},
	}
}

func create401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/department",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &departmentapp.NewDepartment{Label: "Unauthorized"},
		},
	}
}

func createDuplicate409(sd seedData) []apitest.Table {
	existing := sd.Departments[0]

	return []apitest.Table{
		{
			Name:       "duplicate-label",
			URL:        "/department",
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &departmentapp.NewDepartment{Label: existing.Label.String()},
		},
	}
}
