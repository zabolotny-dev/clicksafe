package employeeapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/domain/employeeapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func query200(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/employee",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func queryByID200(sd seedData) []apitest.Table {
	emp := sd.Employees[0]

	return []apitest.Table{
		{
			Name:       "seeded",
			URL:        fmt.Sprintf("/employee/%s", emp.ID),
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &employeeapp.Employee{},
			ExpResp: &employeeapp.Employee{
				ID:        emp.ID,
				FirstName: emp.FirstName.String(),
				LastName:  emp.LastName.String(),
				Email:     emp.Email.Address,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(employeeapp.Employee{}, "Phone", "Attributes", "DepartmentID"))
			},
		},
	}
}

func queryByID404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/employee/%s", uuid.New()),
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
