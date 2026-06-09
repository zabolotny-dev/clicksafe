package employeeapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/employeeapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func update200(sd seedData) []apitest.Table {
	emp := sd.Employees[0]
	newFirst := "UpdatedFirst"

	return []apitest.Table{
		{
			Name:       "change-name",
			URL:        fmt.Sprintf("/employee/%s", emp.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &employeeapp.UpdateEmployee{FirstName: &newFirst},
			GotResp:    &employeeapp.Employee{},
			ExpResp: &employeeapp.Employee{
				ID:        emp.ID,
				FirstName: newFirst,
				LastName:  emp.LastName.String(),
				Email:     emp.Email.Address,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(employeeapp.Employee{}, "Phone", "Attributes", "DepartmentID"))
			},
		},
	}
}
