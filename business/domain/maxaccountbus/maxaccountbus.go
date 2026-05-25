package maxaccountbus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

var (
	ErrInvalidStatus   = errors.New("invalid max account status")
	ErrAccountNotFound = errors.New("max account not found")
	ErrAdapterFailed   = errors.New("max adapter failed")
)

type Storer interface {
	Upsert(ctx context.Context, account Account) (Account, error)
	UpdateLabel(ctx context.Context, account Account) (Account, error)
	Query(ctx context.Context, filter QueryFilter, page page.Page) ([]Account, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (Account, error)
	QueryByAdapterID(ctx context.Context, id uuid.UUID) (Account, error)
	Delete(ctx context.Context, account Account) error
}

type Adapter interface {
	BeginLogin(ctx context.Context, phone string) (LoginAttempt, error)
	ConfirmLogin(ctx context.Context, attemptID, code, password string) (LoginResult, error)
	ConfirmPassword(ctx context.Context, attemptID, password string) (LoginResult, error)
	ListAccounts(ctx context.Context) ([]Account, error)
	StartAccount(ctx context.Context, adapterID uuid.UUID) (Account, error)
	StopAccount(ctx context.Context, adapterID uuid.UUID) (Account, error)
	DeleteAccount(ctx context.Context, adapterID uuid.UUID) error
}

type Business struct {
	storer  Storer
	adapter Adapter
}

func NewBusiness(storer Storer, adapter Adapter) *Business {
	return &Business{storer: storer, adapter: adapter}
}

func (b *Business) BeginLogin(ctx context.Context, nl BeginLogin) (LoginAttempt, error) {
	attempt, err := b.adapter.BeginLogin(ctx, nl.Phone)
	if err != nil {
		return LoginAttempt{}, fmt.Errorf("beginlogin: %w", err)
	}
	attempt.Label = nl.Label
	return attempt, nil
}

func (b *Business) ConfirmLogin(ctx context.Context, cl ConfirmLogin) (LoginResult, error) {
	result, err := b.adapter.ConfirmLogin(ctx, cl.AttemptID, cl.Code, cl.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("confirmlogin: %w", err)
	}
	return b.persistLoginResult(ctx, result, cl.Label)
}

func (b *Business) ConfirmPassword(ctx context.Context, cp ConfirmPassword) (LoginResult, error) {
	result, err := b.adapter.ConfirmPassword(ctx, cp.AttemptID, cp.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("confirmpassword: %w", err)
	}
	return b.persistLoginResult(ctx, result, cp.Label)
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, page page.Page) ([]Account, error) {
	remote, err := b.adapter.ListAccounts(ctx)
	if err == nil {
		for _, account := range remote {
			merged, mergeErr := b.mergeRemoteAccount(ctx, account)
			if mergeErr != nil {
				return nil, fmt.Errorf("query: sync account[%s]: %w", account.AdapterID, mergeErr)
			}

			if _, upsertErr := b.storer.Upsert(ctx, merged); upsertErr != nil {
				return nil, fmt.Errorf("query: sync account[%s]: %w", account.AdapterID, upsertErr)
			}
		}
	}

	accounts, storeErr := b.storer.Query(ctx, filter, page)
	if storeErr != nil {
		return nil, fmt.Errorf("query: %w", storeErr)
	}
	if err != nil {
		return accounts, fmt.Errorf("query: %w", err)
	}
	return accounts, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Account, error) {
	account, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Account{}, fmt.Errorf("querybyid: accountID[%s]: %w", id, err)
	}
	return account, nil
}

func (b *Business) Update(ctx context.Context, account Account, up UpdateAccount) (Account, error) {
	account.Label = up.Label

	updated, err := b.storer.UpdateLabel(ctx, account)
	if err != nil {
		return Account{}, fmt.Errorf("update: %w", err)
	}
	return updated, nil
}

func (b *Business) Start(ctx context.Context, account Account) (Account, error) {
	remote, err := b.adapter.StartAccount(ctx, account.AdapterID)
	if err != nil {
		return Account{}, fmt.Errorf("start: %w", err)
	}
	saved, err := b.storer.Upsert(ctx, mergeLocalAccount(account, remote))
	if err != nil {
		return Account{}, fmt.Errorf("start: %w", err)
	}
	return saved, nil
}

func (b *Business) Stop(ctx context.Context, account Account) (Account, error) {
	remote, err := b.adapter.StopAccount(ctx, account.AdapterID)
	if err != nil {
		return Account{}, fmt.Errorf("stop: %w", err)
	}
	saved, err := b.storer.Upsert(ctx, mergeLocalAccount(account, remote))
	if err != nil {
		return Account{}, fmt.Errorf("stop: %w", err)
	}
	return saved, nil
}

func (b *Business) Delete(ctx context.Context, account Account) error {
	if err := b.adapter.DeleteAccount(ctx, account.AdapterID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if err := b.storer.Delete(ctx, account); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (b *Business) persistLoginResult(ctx context.Context, result LoginResult, localLabel string) (LoginResult, error) {
	result.Attempt.Label = firstNonBlank(localLabel, result.Attempt.Label)

	if result.Account == nil {
		if result.Attempt.Label == "" {
			result.Attempt.Label = defaultAccountLabel(result.Attempt.Phone)
		}
		return result, nil
	}

	if localLabel == "" {
		localLabel = initialLocalLabel(*result.Account)
	}
	result.Account.Label = localLabel
	result.Attempt.Label = localLabel

	saved, err := b.storer.Upsert(ctx, *result.Account)
	if err != nil {
		return LoginResult{}, fmt.Errorf("persistloginresult: %w", err)
	}
	result.Account = &saved
	return result, nil
}

func (b *Business) mergeRemoteAccount(ctx context.Context, remote Account) (Account, error) {
	local, err := b.storer.QueryByAdapterID(ctx, remote.AdapterID)
	if err == nil {
		return mergeLocalAccount(local, remote), nil
	}
	if !errors.Is(err, ErrAccountNotFound) {
		return Account{}, err
	}

	remote.Label = initialLocalLabel(remote)
	return remote, nil
}

func mergeLocalAccount(local Account, remote Account) Account {
	remote.ID = local.ID
	remote.Label = local.Label
	return remote
}

func initialLocalLabel(account Account) string {
	if candidate := strings.TrimSpace(account.Label); candidate != "" {
		if parsed, err := label.Parse(candidate); err == nil {
			return parsed.String()
		}
	}
	return defaultAccountLabel(account.Phone)
}

func defaultAccountLabel(phone string) string {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return "Max account"
	}
	return "Max " + strings.TrimPrefix(trimmed, "+")
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
