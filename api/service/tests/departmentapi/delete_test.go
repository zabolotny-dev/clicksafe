package departmentapi_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func delete204(db *dbtest.Database, sd seedData) []apitest.Table {
	dep, err := db.BusDomain.Department.Save(context.Background(), departmentbus.NewDepartment{
		Label: label.MustParse("To Be Deleted"),
	})
	if err != nil {
		panic("delete_test seed: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "existing",
			URL:        fmt.Sprintf("/department/%s", dep.ID),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func delete404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/department/%s", uuid.New()),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
