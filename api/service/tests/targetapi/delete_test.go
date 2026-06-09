package targetapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

func deleteByID204(db *dbtest.Database, sd seedData) []apitest.Table {
	ctx := context.Background()

	emp, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName: name.MustParse("Delete"),
		LastName:  name.MustParse("Me"),
		Email:     mail.Address{Address: "delete.me@example.com"},
	})
	if err != nil {
		panic("delete_test seed employee: " + err.Error())
	}

	tgt, err := db.BusDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: sd.Campaign.ID,
		EmployeeID: emp.ID,
	})
	if err != nil {
		panic("delete_test seed target: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "existing",
			URL:        fmt.Sprintf("/target/%s", tgt.ID),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func deleteByID404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/target/%s", uuid.New()),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func deleteByCampaign204(db *dbtest.Database, sd seedData) []apitest.Table {
	ctx := context.Background()

	emp, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
		FirstName: name.MustParse("Camp"),
		LastName:  name.MustParse("Delete"),
		Email:     mail.Address{Address: "camp.delete@example.com"},
	})
	if err != nil {
		panic("deleteByCampaign seed employee: " + err.Error())
	}

	_, err = db.BusDomain.Target.Save(ctx, campaignbus.NewTarget{
		CampaignID: sd.Campaign.ID,
		EmployeeID: emp.ID,
	})
	if err != nil {
		panic("deleteByCampaign seed target: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "existing",
			URL:        fmt.Sprintf("/target/campaign/%s", sd.Campaign.ID),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func deleteByCampaign404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/target/campaign/%s", uuid.New()),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
