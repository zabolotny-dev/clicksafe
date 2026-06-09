package departmentapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/departmentapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func update200(sd seedData) []apitest.Table {
	dep := sd.Departments[0]
	newLabel := "Updated Department"

	return []apitest.Table{
		{
			Name:       "change-label",
			URL:        fmt.Sprintf("/department/%s", dep.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &departmentapp.UpdateDepartment{Label: &newLabel},
			GotResp:    &departmentapp.Department{},
			ExpResp:    &departmentapp.Department{ID: dep.ID, Label: newLabel},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(departmentapp.Department{}, "Attributes"))
			},
		},
	}
}

func update401(sd seedData) []apitest.Table {
	dep := sd.Departments[0]
	newLabel := "Should Fail"

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/department/%s", dep.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &departmentapp.UpdateDepartment{Label: &newLabel},
		},
	}
}
