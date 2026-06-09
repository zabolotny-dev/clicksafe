package messageapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/domain/messageapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func query200(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/message",
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}

func queryByID200(sd seedData) []apitest.Table {
	msg := sd.Messages[0]

	return []apitest.Table{
		{
			Name:       "seeded",
			URL:        fmt.Sprintf("/message/%s", msg.ID),
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			GotResp:    &messageapp.Message{},
			ExpResp: &messageapp.Message{
				ID:        msg.ID,
				Type:      msg.Type.String(),
				Label:     msg.Label.String(),
				FromEmail: msg.FromEmail.Address,
				FromName:  msg.FromName.String(),
				Subject:   msg.Subject.String(),
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp, cmpopts.IgnoreFields(messageapp.Message{},
					"HtmlBodyID", "TextBodyID", "MaxAccountID", "AttachmentIDs"))
			},
		},
	}
}

func queryByID404(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/message/%s", uuid.New()),
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
		},
	}
}
