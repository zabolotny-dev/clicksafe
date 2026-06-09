package employeeapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

func delete204(db *dbtest.Database, sd seedData) []apitest.Table {
	emp, err := db.BusDomain.Employee.Save(context.Background(), employeebus.NewEmployee{
		FirstName: name.MustParse("To"),
		LastName:  name.MustParse("Delete"),
		Email:     mail.Address{Address: "todelete@example.com"},
	})
	if err != nil {
		panic("delete_test seed: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "existing",
			URL:        fmt.Sprintf("/employee/%s", emp.ID),
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
			URL:        fmt.Sprintf("/employee/%s", uuid.New()),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
