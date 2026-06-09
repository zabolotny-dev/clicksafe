package authapi_test

import (
	"net/http"

	"github.com/zabolotny-dev/clicksafe/app/domain/authapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
)

func login200(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "valid-credentials",
			URL:        "/login",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &authapp.Login{
				Login:    sd.RawLogin,
				Password: rawPassword,
			},
		},
	}
}

func login401(sd seedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "wrong-password",
			URL:        "/login",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &authapp.Login{
				Login:    sd.RawLogin,
				Password: "WrongPassword999!",
			},
		},
	}
}
