package messageapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/zabolotny-dev/clicksafe/app/domain/messageapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func create201(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/message",
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input: &messageapp.NewMessage{
				Label:     "New Message",
				FromEmail: "sender@example.com",
				FromName:  "Sender",
				Subject:   "Hello World",
			},
			GotResp: &messageapp.Message{},
			ExpResp: &messageapp.Message{
				Type:      "EMAIL",
				Label:     "New Message",
				FromEmail: "sender@example.com",
				FromName:  "Sender",
				Subject:   "Hello World",
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*messageapp.Message)
				e := exp.(*messageapp.Message)
				e.ID = g.ID
				return cmp.Diff(g, e, cmpopts.IgnoreFields(messageapp.Message{},
					"HtmlBodyID", "TextBodyID", "MaxAccountID", "AttachmentIDs"))
			},
		},
	}
}

func create400(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-label",
			URL:        "/message",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &messageapp.NewMessage{FromEmail: "a@b.com"},
		},
		{
			Name:       "invalid-email",
			URL:        "/message",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Token:      sd.Admins[0].RawToken,
			CSRFToken:  sd.Admins[0].CSRFToken,
			Input:      &messageapp.NewMessage{Label: "Bad Email", FromEmail: "not-an-email"},
		},
	}
}

func create401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/message",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &messageapp.NewMessage{
				Label:     "Unauthorized",
				FromEmail: "anon@example.com",
			},
		},
	}
}
