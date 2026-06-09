package targetapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/targetapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func schedule200(sd seedData) []apitest.Table {
	tgt := sd.Target
	scheduledAt := time.Now().UTC().Add(24 * time.Hour)

	return []apitest.Table{
		{
			Name:       "set-schedule",
			URL:        fmt.Sprintf("/target/%s/schedule", tgt.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &targetapp.UpdateSchedule{ScheduledAt: scheduledAt},
			GotResp:    &targetapp.Target{},
			ExpResp: &targetapp.Target{
				ID:         tgt.ID,
				EmployeeID: tgt.EmployeeID,
				CampaignID: tgt.CampaignID,
				Status:     tgt.Status.String(),
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(targetapp.Target{},
					"Token", "ScheduledAt", "CreatedAt"))
			},
		},
	}
}

func schedule401(sd seedData) []apitest.Table {
	tgt := sd.Target
	scheduledAt := time.Now().UTC().Add(24 * time.Hour)

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/target/%s/schedule", tgt.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &targetapp.UpdateSchedule{ScheduledAt: scheduledAt},
		},
	}
}
