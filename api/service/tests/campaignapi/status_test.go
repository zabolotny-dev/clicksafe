package campaignapi_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func newDraft(db *dbtest.Database) campaignbus.Campaign {
	cmp, err := db.BusDomain.Campaign.Save(context.Background(), campaignbus.NewCampaign{
		Label: label.MustParse("Status Test Campaign " + uuid.New().String()[:8]),
	})
	if err != nil {
		panic("status_test: " + err.Error())
	}
	return cmp
}

// start on empty Draft → missing required fields → 400.
func startDraft400(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	cmp := newDraft(db)
	return []apitest.Table{
		{
			Name:       "missing-fields",
			URL:        fmt.Sprintf("/campaign/%s/start", cmp.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

// start on Completed → invalid transition → 400.
func startCompleted400(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	// Seed a campaign and cancel it so it reaches a terminal state.
	cmp := newDraft(db)
	canceled, err := db.BusDomain.Campaign.Cancel(context.Background(), cmp)
	if err != nil {
		panic("status_test cancel: " + err.Error())
	}

	return []apitest.Table{
		{
			Name:       "canceled-start-invalid",
			URL:        fmt.Sprintf("/campaign/%s/start", canceled.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

// pause on Draft → invalid transition → 400.
func pauseDraft400(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	cmp := newDraft(db)
	return []apitest.Table{
		{
			Name:       "draft-pause-invalid",
			URL:        fmt.Sprintf("/campaign/%s/pause", cmp.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

// cancel on Draft → 200 Canceled.
func cancelDraft200(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	cmp := newDraft(db)
	return []apitest.Table{
		{
			Name:       "draft-to-canceled",
			URL:        fmt.Sprintf("/campaign/%s/cancel", cmp.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

// cancel on already Canceled → invalid transition → 400.
func cancelCanceled400(db *dbtest.Database, sd apitest.SeedData) []apitest.Table {
	cmp := newDraft(db)
	canceled, _ := db.BusDomain.Campaign.Cancel(context.Background(), cmp)
	return []apitest.Table{
		{
			Name:       "double-cancel",
			URL:        fmt.Sprintf("/campaign/%s/cancel", canceled.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
