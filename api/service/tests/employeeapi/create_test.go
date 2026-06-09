package employeeapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/employeeapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/employee",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &employeeapp.NewEmployee{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
			},
			GotResp: &employeeapp.Employee{},
			ExpResp: &employeeapp.Employee{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*employeeapp.Employee)
				e := exp.(*employeeapp.Employee)
				e.ID = g.ID
				return cmp.Diff(g, e, cmpopts.IgnoreFields(employeeapp.Employee{}, "Phone", "Attributes", "DepartmentID"))
			},
		},
	}
}

func create400(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-email",
			URL:        "/employee",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &employeeapp.NewEmployee{FirstName: "No", LastName: "Email"},
		},
		{
			Name:       "invalid-email",
			URL:        "/employee",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &employeeapp.NewEmployee{FirstName: "Bad", LastName: "Email", Email: "not-an-email"},
		},
	}
}

func create401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/employee",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &employeeapp.NewEmployee{
				FirstName: "Anon",
				LastName:  "User",
				Email:     "anon@example.com",
			},
		},
	}
}
