package departmentapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/domain/departmentapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func query200(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/department",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func queryByID200(sd seedData) []apitest.Table {
	dep := sd.Departments[0]

	return []apitest.Table{
		{
			Name:       "seeded",
			URL:        fmt.Sprintf("/department/%s", dep.ID),
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &departmentapp.Department{},
			ExpResp:    &departmentapp.Department{ID: dep.ID, Label: dep.Label.String()},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(departmentapp.Department{}, "Attributes"))
			},
		},
	}
}

func queryByID404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/department/%s", uuid.New()),
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
