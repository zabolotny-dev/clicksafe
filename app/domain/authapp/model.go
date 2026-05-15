package authapp

import (
	"time"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
)

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResult struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CSRFToken string    `json:"csrf_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Me struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CSRFToken string    `json:"csrf_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func toBusLoginData(r Login) (authbus.LoginData, error) {
	var errors errs.FieldErrors

	log, err := login.Parse(r.Login)
	if err != nil {
		errors.Add("login", err)
	}

	pass, err := password.Parse(r.Password)
	if err != nil {
		errors.Add("password", err)
	}

	if len(errors) > 0 {
		return authbus.LoginData{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return authbus.LoginData{
		Login:    log,
		Password: pass,
	}, nil
}
