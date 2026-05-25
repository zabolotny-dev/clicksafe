package maxaccountapp

import (
	"strings"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

type Account struct {
	ID        uuid.UUID `json:"id"`
	AdapterID uuid.UUID `json:"adapter_id"`
	Phone     string    `json:"phone"`
	Label     string    `json:"label"`
	Status    string    `json:"status"`
	MaxUserID string    `json:"max_user_id"`
	LastError string    `json:"last_error"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type LoginAttempt struct {
	ID        string `json:"id"`
	Phone     string `json:"phone"`
	Label     string `json:"label"`
	Status    string `json:"status"`
	Error     string `json:"error"`
	ExpiresAt string `json:"expires_at"`
}

type LoginResult struct {
	Attempt LoginAttempt `json:"attempt"`
	Account *Account     `json:"account"`
}

type BeginLoginRequest struct {
	Phone string `json:"phone"`
	Label string `json:"label"`
}

type ConfirmLoginRequest struct {
	Code     string `json:"code"`
	Password string `json:"password"`
	Label    string `json:"label"`
}

type ConfirmPasswordRequest struct {
	Password string `json:"password"`
	Label    string `json:"label"`
}

type UpdateAccountRequest struct {
	Label string `json:"label"`
}

func toBusBeginLogin(req BeginLoginRequest) (maxaccountbus.BeginLogin, error) {
	var fieldErrors errs.FieldErrors

	ph, err := phone.Parse(req.Phone)
	if err != nil {
		fieldErrors.Add("phone", err)
	}

	lbl := strings.TrimSpace(req.Label)
	if lbl == "" && err == nil {
		lbl = "Max " + ph.String()[1:]
	}
	parsedLabel, err := label.Parse(lbl)
	if err != nil {
		fieldErrors.Add("label", err)
	}

	if len(fieldErrors) > 0 {
		return maxaccountbus.BeginLogin{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return maxaccountbus.BeginLogin{
		Phone: ph.String(),
		Label: parsedLabel.String(),
	}, nil
}

func toBusConfirmLogin(req ConfirmLoginRequest, attemptID string) (maxaccountbus.ConfirmLogin, error) {
	var fieldErrors errs.FieldErrors

	lbl := strings.TrimSpace(req.Label)
	if lbl != "" {
		parsedLabel, err := label.Parse(lbl)
		if err != nil {
			fieldErrors.Add("label", err)
		} else {
			lbl = parsedLabel.String()
		}
	}

	if len(fieldErrors) > 0 {
		return maxaccountbus.ConfirmLogin{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return maxaccountbus.ConfirmLogin{
		AttemptID: attemptID,
		Code:      req.Code,
		Password:  req.Password,
		Label:     lbl,
	}, nil
}

func toBusConfirmPassword(req ConfirmPasswordRequest, attemptID string) (maxaccountbus.ConfirmPassword, error) {
	var fieldErrors errs.FieldErrors

	lbl := strings.TrimSpace(req.Label)
	if lbl != "" {
		parsedLabel, err := label.Parse(lbl)
		if err != nil {
			fieldErrors.Add("label", err)
		} else {
			lbl = parsedLabel.String()
		}
	}

	if len(fieldErrors) > 0 {
		return maxaccountbus.ConfirmPassword{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return maxaccountbus.ConfirmPassword{
		AttemptID: attemptID,
		Password:  req.Password,
		Label:     lbl,
	}, nil
}

func toBusUpdateAccount(req UpdateAccountRequest) (maxaccountbus.UpdateAccount, error) {
	var fieldErrors errs.FieldErrors

	parsedLabel, err := label.Parse(strings.TrimSpace(req.Label))
	if err != nil {
		fieldErrors.Add("label", err)
	}

	if len(fieldErrors) > 0 {
		return maxaccountbus.UpdateAccount{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return maxaccountbus.UpdateAccount{
		Label: parsedLabel.String(),
	}, nil
}

func toAppAccount(account maxaccountbus.Account) Account {
	return Account{
		ID:        account.ID,
		AdapterID: account.AdapterID,
		Phone:     account.Phone,
		Label:     account.Label,
		Status:    account.Status.String(),
		MaxUserID: account.MaxUserID,
		LastError: account.LastError,
		CreatedAt: account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: account.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppAccounts(accounts []maxaccountbus.Account) []Account {
	result := make([]Account, len(accounts))
	for i, account := range accounts {
		result[i] = toAppAccount(account)
	}
	return result
}

func toAppLoginAttempt(attempt maxaccountbus.LoginAttempt) LoginAttempt {
	return LoginAttempt{
		ID:        attempt.ID,
		Phone:     attempt.Phone,
		Label:     attempt.Label,
		Status:    attempt.Status,
		Error:     attempt.Error,
		ExpiresAt: attempt.ExpiresAt,
	}
}

func toAppLoginResult(result maxaccountbus.LoginResult) LoginResult {
	var account *Account
	if result.Account != nil {
		appAccount := toAppAccount(*result.Account)
		account = &appAccount
	}
	return LoginResult{
		Attempt: toAppLoginAttempt(result.Attempt),
		Account: account,
	}
}
