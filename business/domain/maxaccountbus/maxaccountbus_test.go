package maxaccountbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

// =============================================================================
// Stubs

type maxStorerStub struct {
	data       map[uuid.UUID]maxaccountbus.Account
	adapterMap map[uuid.UUID]maxaccountbus.Account // keyed by AdapterID
	upsertErr  error
	queryErr   error
}

func newMaxStorerStub() *maxStorerStub {
	return &maxStorerStub{
		data:       make(map[uuid.UUID]maxaccountbus.Account),
		adapterMap: make(map[uuid.UUID]maxaccountbus.Account),
	}
}

func (s *maxStorerStub) Upsert(_ context.Context, acc maxaccountbus.Account) (maxaccountbus.Account, error) {
	if s.upsertErr != nil {
		return maxaccountbus.Account{}, s.upsertErr
	}
	if acc.ID == (uuid.UUID{}) {
		acc.ID = uuid.New()
	}
	s.data[acc.ID] = acc
	s.adapterMap[acc.AdapterID] = acc
	return acc, nil
}

func (s *maxStorerStub) UpdateLabel(_ context.Context, acc maxaccountbus.Account) (maxaccountbus.Account, error) {
	if s.upsertErr != nil {
		return maxaccountbus.Account{}, s.upsertErr
	}
	s.data[acc.ID] = acc
	s.adapterMap[acc.AdapterID] = acc
	return acc, nil
}

func (s *maxStorerStub) Query(_ context.Context, _ maxaccountbus.QueryFilter, _ page.Page) ([]maxaccountbus.Account, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	var result []maxaccountbus.Account
	for _, a := range s.data {
		result = append(result, a)
	}
	return result, nil
}

func (s *maxStorerStub) Count(_ context.Context, _ maxaccountbus.QueryFilter) (int, error) {
	return len(s.data), nil
}

func (s *maxStorerStub) QueryByID(_ context.Context, id uuid.UUID) (maxaccountbus.Account, error) {
	a, ok := s.data[id]
	if !ok {
		return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
	}
	return a, nil
}

func (s *maxStorerStub) QueryByAdapterID(_ context.Context, id uuid.UUID) (maxaccountbus.Account, error) {
	a, ok := s.adapterMap[id]
	if !ok {
		return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
	}
	return a, nil
}

func (s *maxStorerStub) Delete(_ context.Context, acc maxaccountbus.Account) error {
	delete(s.data, acc.ID)
	delete(s.adapterMap, acc.AdapterID)
	return nil
}

type maxAdapterStub struct {
	attempt    maxaccountbus.LoginAttempt
	result     maxaccountbus.LoginResult
	accounts   []maxaccountbus.Account
	startAcc   maxaccountbus.Account
	stopAcc    maxaccountbus.Account
	err        error
	deletedIDs []uuid.UUID
}

func (a *maxAdapterStub) BeginLogin(_ context.Context, phone string) (maxaccountbus.LoginAttempt, error) {
	if a.err != nil {
		return maxaccountbus.LoginAttempt{}, a.err
	}
	attempt := a.attempt
	attempt.Phone = phone
	return attempt, nil
}

func (a *maxAdapterStub) ConfirmLogin(_ context.Context, _, _, _ string) (maxaccountbus.LoginResult, error) {
	if a.err != nil {
		return maxaccountbus.LoginResult{}, a.err
	}
	return a.result, nil
}

func (a *maxAdapterStub) ConfirmPassword(_ context.Context, _, _ string) (maxaccountbus.LoginResult, error) {
	if a.err != nil {
		return maxaccountbus.LoginResult{}, a.err
	}
	return a.result, nil
}

func (a *maxAdapterStub) ListAccounts(_ context.Context) ([]maxaccountbus.Account, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.accounts, nil
}

func (a *maxAdapterStub) StartAccount(_ context.Context, adapterID uuid.UUID) (maxaccountbus.Account, error) {
	if a.err != nil {
		return maxaccountbus.Account{}, a.err
	}
	acc := a.startAcc
	acc.AdapterID = adapterID
	return acc, nil
}

func (a *maxAdapterStub) StopAccount(_ context.Context, adapterID uuid.UUID) (maxaccountbus.Account, error) {
	if a.err != nil {
		return maxaccountbus.Account{}, a.err
	}
	acc := a.stopAcc
	acc.AdapterID = adapterID
	return acc, nil
}

func (a *maxAdapterStub) DeleteAccount(_ context.Context, id uuid.UUID) error {
	if a.err != nil {
		return a.err
	}
	a.deletedIDs = append(a.deletedIDs, id)
	return nil
}

// =============================================================================
// BeginLogin tests

func TestBeginLogin_SetsPhoneFromArgument(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{
		attempt: maxaccountbus.LoginAttempt{ID: "attempt-1"},
	}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	got, err := bus.BeginLogin(context.Background(), maxaccountbus.BeginLogin{
		Phone: "+79991234567",
		Label: "My account",
	})
	if err != nil {
		t.Fatalf("BeginLogin returned error: %v", err)
	}
	if got.Phone != "+79991234567" {
		t.Errorf("Phone = %q, want %q", got.Phone, "+79991234567")
	}
}

func TestBeginLogin_SetsLabelFromArgument(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{
		attempt: maxaccountbus.LoginAttempt{ID: "attempt-1"},
	}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	got, err := bus.BeginLogin(context.Background(), maxaccountbus.BeginLogin{
		Phone: "+79991234567",
		Label: "My VIP account",
	})
	if err != nil {
		t.Fatalf("BeginLogin returned error: %v", err)
	}
	if got.Label != "My VIP account" {
		t.Errorf("Label = %q, want %q", got.Label, "My VIP account")
	}
}

func TestBeginLogin_AdapterError_Propagates(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{err: errors.New("adapter failure")}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	_, err := bus.BeginLogin(context.Background(), maxaccountbus.BeginLogin{Phone: "+79991234567"})
	if err == nil {
		t.Fatal("expected error from adapter, got nil")
	}
}

// =============================================================================
// ConfirmLogin tests

func TestConfirmLogin_WithAccount_PersistsAndReturns(t *testing.T) {
	t.Parallel()

	adapterID := uuid.New()
	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: adapterID,
		Phone:     "+79991234567",
		Label:     "My account",
		Status:    maxaccountbus.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter := &maxAdapterStub{
		result: maxaccountbus.LoginResult{
			Attempt: maxaccountbus.LoginAttempt{ID: "attempt-1", Phone: "+79991234567"},
			Account: &acc,
		},
	}
	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	result, err := bus.ConfirmLogin(context.Background(), maxaccountbus.ConfirmLogin{
		AttemptID: "attempt-1",
		Code:      "12345",
		Password:  "pass",
		Label:     "My account",
	})
	if err != nil {
		t.Fatalf("ConfirmLogin returned error: %v", err)
	}
	if result.Account == nil {
		t.Fatal("expected non-nil Account in result")
	}
	if _, exists := store.adapterMap[adapterID]; !exists {
		t.Error("Account was not persisted to storer")
	}
}

func TestConfirmLogin_NoAccount_ReturnsAttemptOnly(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{
		result: maxaccountbus.LoginResult{
			Attempt: maxaccountbus.LoginAttempt{
				ID:    "attempt-1",
				Phone: "+79991234567",
			},
			Account: nil, // Нет готового аккаунта — требует подтверждения пароля
		},
	}

	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	result, err := bus.ConfirmLogin(context.Background(), maxaccountbus.ConfirmLogin{
		AttemptID: "attempt-1",
		Code:      "12345",
		Password:  "pass",
	})
	if err != nil {
		t.Fatalf("ConfirmLogin returned error: %v", err)
	}
	if result.Account != nil {
		t.Error("expected nil Account when login not complete")
	}
	if result.Attempt.ID != "attempt-1" {
		t.Errorf("Attempt.ID = %q, want attempt-1", result.Attempt.ID)
	}
	if len(store.data) != 0 {
		t.Errorf("storer must have 0 accounts, got %d", len(store.data))
	}
}

func TestConfirmLogin_AdapterError_Propagates(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{err: errors.New("adapter failure")}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	_, err := bus.ConfirmLogin(context.Background(), maxaccountbus.ConfirmLogin{})
	if err == nil {
		t.Fatal("expected error from adapter, got nil")
	}
}

// =============================================================================
// Query tests

func TestQuery_WithAdapterSuccess_SyncsAccounts(t *testing.T) {
	t.Parallel()

	adapterID := uuid.New()
	remoteAcc := maxaccountbus.Account{
		AdapterID: adapterID,
		Phone:     "+79991234567",
		Status:    maxaccountbus.StatusActive,
	}

	adapter := &maxAdapterStub{accounts: []maxaccountbus.Account{remoteAcc}}
	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	accounts, err := bus.Query(context.Background(), maxaccountbus.QueryFilter{}, page.MustParse("1", "10"))
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	// После синхронизации аккаунт должен быть в сторе
	if _, exists := store.adapterMap[adapterID]; !exists {
		t.Error("remote account was not upserted into storer")
	}
	if len(accounts) == 0 {
		t.Error("Query must return accounts after sync")
	}
}

func TestQuery_AdapterFails_StillReturnsLocalData(t *testing.T) {
	t.Parallel()

	store := newMaxStorerStub()
	localID := uuid.New()
	store.data[localID] = maxaccountbus.Account{
		ID:        localID,
		Phone:     "+79991234567",
		Status:    maxaccountbus.StatusActive,
	}

	adapter := &maxAdapterStub{err: errors.New("network error")}
	bus := maxaccountbus.NewBusiness(store, adapter)

	accounts, err := bus.Query(context.Background(), maxaccountbus.QueryFilter{}, page.MustParse("1", "10"))
	// Ошибка адаптера оборачивается, но данные из стора возвращаются
	if err == nil {
		t.Fatal("expected error when adapter fails")
	}
	if len(accounts) == 0 {
		t.Error("Query must return local accounts even when adapter fails")
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsAccount(t *testing.T) {
	t.Parallel()

	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, &maxAdapterStub{})

	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: uuid.New(),
		Phone:     "+79991234567",
	}
	store.data[acc.ID] = acc

	got, err := bus.QueryByID(context.Background(), acc.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != acc.ID {
		t.Errorf("ID = %v, want %v", got.ID, acc.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), &maxAdapterStub{})

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, maxaccountbus.ErrAccountNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, maxaccountbus.ErrAccountNotFound)
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, &maxAdapterStub{})

	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: uuid.New(),
		Phone:     "+79991234567",
		Label:     "Old Label",
	}
	store.data[acc.ID] = acc
	store.adapterMap[acc.AdapterID] = acc

	updated, err := bus.Update(context.Background(), acc, maxaccountbus.UpdateAccount{
		Label: "New Label",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != "New Label" {
		t.Errorf("Label = %q, want New Label", updated.Label)
	}
}

// =============================================================================
// Start tests

func TestStart_CallsAdapterAndSavesAccount(t *testing.T) {
	t.Parallel()

	adapterID := uuid.New()
	adapter := &maxAdapterStub{
		startAcc: maxaccountbus.Account{
			Phone:  "+79991234567",
			Status: maxaccountbus.StatusActive,
		},
	}
	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: adapterID,
		Label:     "My Account",
		Phone:     "+79991234567",
	}
	store.data[acc.ID] = acc
	store.adapterMap[acc.AdapterID] = acc

	result, err := bus.Start(context.Background(), acc)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if result.Status != maxaccountbus.StatusActive {
		t.Errorf("Status = %v, want %v", result.Status, maxaccountbus.StatusActive)
	}
}

func TestStart_AdapterError_Propagates(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{err: errors.New("adapter error")}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	_, err := bus.Start(context.Background(), maxaccountbus.Account{AdapterID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from adapter, got nil")
	}
}

// =============================================================================
// Stop tests

func TestStop_CallsAdapterAndSavesAccount(t *testing.T) {
	t.Parallel()

	adapterID := uuid.New()
	adapter := &maxAdapterStub{
		stopAcc: maxaccountbus.Account{
			Phone:  "+79991234567",
			Status: maxaccountbus.StatusDisconnected,
		},
	}
	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: adapterID,
		Label:     "My Account",
		Phone:     "+79991234567",
		Status:    maxaccountbus.StatusActive,
	}
	store.data[acc.ID] = acc
	store.adapterMap[acc.AdapterID] = acc

	result, err := bus.Stop(context.Background(), acc)
	if err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if result.Status != maxaccountbus.StatusDisconnected {
		t.Errorf("Status = %v, want %v", result.Status, maxaccountbus.StatusDisconnected)
	}
}

// =============================================================================
// Delete tests

func TestDelete_CallsAdapterThenStorer(t *testing.T) {
	t.Parallel()

	adapterID := uuid.New()
	adapter := &maxAdapterStub{}
	store := newMaxStorerStub()
	bus := maxaccountbus.NewBusiness(store, adapter)

	acc := maxaccountbus.Account{
		ID:        uuid.New(),
		AdapterID: adapterID,
		Phone:     "+79991234567",
	}
	store.data[acc.ID] = acc
	store.adapterMap[acc.AdapterID] = acc

	if err := bus.Delete(context.Background(), acc); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, exists := store.data[acc.ID]; exists {
		t.Error("Account still exists in storer after Delete")
	}

	found := false
	for _, id := range adapter.deletedIDs {
		if id == adapterID {
			found = true
			break
		}
	}
	if !found {
		t.Error("adapter DeleteAccount was not called with correct adapterID")
	}
}

func TestDelete_AdapterError_Propagates(t *testing.T) {
	t.Parallel()

	adapter := &maxAdapterStub{err: errors.New("adapter error")}
	bus := maxaccountbus.NewBusiness(newMaxStorerStub(), adapter)

	err := bus.Delete(context.Background(), maxaccountbus.Account{AdapterID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from adapter, got nil")
	}
}
