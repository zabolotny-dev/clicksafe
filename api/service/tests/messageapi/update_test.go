package messageapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/messageapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func update200(sd seedData) []apitest.Table {
	msg := sd.Messages[0]
	newLabel := "Updated Message"

	return []apitest.Table{
		{
			Name:       "change-label",
			URL:        fmt.Sprintf("/message/%s", msg.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &messageapp.UpdateMessage{Label: &newLabel},
			GotResp:    &messageapp.Message{},
			ExpResp: &messageapp.Message{
				ID:    msg.ID,
				Label: newLabel,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(messageapp.Message{},
					"Type", "FromEmail", "FromName", "Subject",
					"HtmlBodyID", "TextBodyID", "MaxAccountID", "AttachmentIDs"))
			},
		},
	}
}

func update401(sd seedData) []apitest.Table {
	msg := sd.Messages[0]
	newLabel := "Should Fail"

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/message/%s", msg.ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &messageapp.UpdateMessage{Label: &newLabel},
		},
	}
}
