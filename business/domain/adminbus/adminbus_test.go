package adminbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

// =============================================================================
// Stubs

type adminStorerStub struct {
	saved    adminbus.Admin
	data     map[uuid.UUID]adminbus.Admin
	loginMap map[string]adminbus.Admin
	saveErr  error
	queryErr error
}

func newAdminStorerStub() *adminStorerStub {
	return &adminStorerStub{
		data:     make(map[uuid.UUID]adminbus.Admin),
		loginMap: make(map[string]adminbus.Admin),
	}
}

func (s *adminStorerStub) Save(_ context.Context, adm adminbus.Admin) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = adm
	s.data[adm.ID] = adm
	s.loginMap[adm.Login.String()] = adm
	return nil
}

func (s *adminStorerStub) Update(_ context.Context, adm adminbus.Admin) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = adm
	s.data[adm.ID] = adm
	s.loginMap[adm.Login.String()] = adm
	return nil
}

func (s *adminStorerStub) QueryByID(_ context.Context, id uuid.UUID) (adminbus.Admin, error) {
	if s.queryErr != nil {
		return adminbus.Admin{}, s.queryErr
	}
	adm, ok := s.data[id]
	if !ok {
		return adminbus.Admin{}, adminbus.ErrNotFound
	}
	return adm, nil
}

func (s *adminStorerStub) QueryByLogin(_ context.Context, lgn login.Login) (adminbus.Admin, error) {
	if s.queryErr != nil {
		return adminbus.Admin{}, s.queryErr
	}
	adm, ok := s.loginMap[lgn.String()]
	if !ok {
		return adminbus.Admin{}, adminbus.ErrNotFound
	}
	return adm, nil
}

func (s *adminStorerStub) Query(_ context.Context, _ adminbus.AdminQueryFilter) ([]adminbus.Admin, error) {
	return nil, nil
}

func (s *adminStorerStub) Count(_ context.Context, _ adminbus.AdminQueryFilter) (int, error) {
	return len(s.data), nil
}

// hasherStub записывает пароль с префиксом "hashed:" и сравнивает по тому же правилу.
type hasherStub struct {
	generateErr error
}

func (h *hasherStub) Generate(p string) (string, error) {
	if h.generateErr != nil {
		return "", h.generateErr
	}
	return "hashed:" + p, nil
}

func (h *hasherStub) Compare(p, encodedHash string) (bool, error) {
	return encodedHash == "hashed:"+p, nil
}

// =============================================================================
// Save tests

func TestSave_CreatesAdminWithHashedPassword(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	newAdm := adminbus.NewAdmin{
		FirstName: name.MustParse("Ivan"),
		LastName:  name.MustParse("Petrov"),
		Login:     login.MustParse("ivan.petrov"),
		Password:  password.MustParse("Secret123!@#pass"),
	}

	adm, err := bus.Save(context.Background(), newAdm)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if adm.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if adm.PasswordHash != "hashed:Secret123!@#pass" {
		t.Errorf("PasswordHash = %q, want %q", adm.PasswordHash, "hashed:Secret123!@#pass")
	}
	if adm.FirstName != newAdm.FirstName {
		t.Errorf("FirstName = %v, want %v", adm.FirstName, newAdm.FirstName)
	}
	if adm.Login != newAdm.Login {
		t.Errorf("Login = %v, want %v", adm.Login, newAdm.Login)
	}
	if store.saved.ID != adm.ID {
		t.Error("Admin was not persisted to storer")
	}
}

func TestSave_HasherError_ReturnsError(t *testing.T) {
	t.Parallel()

	hasher := &hasherStub{generateErr: errors.New("bcrypt failed")}
	bus := adminbus.NewBusiness(hasher, newAdminStorerStub())

	_, err := bus.Save(context.Background(), adminbus.NewAdmin{
		Login:    login.MustParse("test.user"),
		Password: password.MustParse("Secret123!@#pass"),
	})
	if err == nil {
		t.Fatal("expected error when hasher fails, got nil")
	}
}

func TestSave_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	store.saveErr = errors.New("db error")
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	_, err := bus.Save(context.Background(), adminbus.NewAdmin{
		Login:    login.MustParse("test.user"),
		Password: password.MustParse("Secret123!@#pass"),
	})
	if err == nil {
		t.Fatal("expected error when store fails, got nil")
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesFirstName(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	// Создаём исходного admin в стабе
	original := adminbus.Admin{
		ID:           uuid.New(),
		FirstName:    name.MustParse("Staryi"),
		LastName:     name.MustParse("Imya"),
		Login:        login.MustParse("old.name"),
		PasswordHash: "hashed:pass",
	}
	store.data[original.ID] = original
	store.loginMap[original.Login.String()] = original

	newFirst := name.MustParse("Novoe")
	updated, err := bus.Update(context.Background(), original.ID, adminbus.UpdateAdmin{
		FirstName: &newFirst,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.FirstName != newFirst {
		t.Errorf("FirstName = %v, want %v", updated.FirstName, newFirst)
	}
	// LastName и Login не должны измениться
	if updated.LastName != original.LastName {
		t.Errorf("LastName changed unexpectedly: got %v, want %v", updated.LastName, original.LastName)
	}
}

func TestUpdate_ChangesPassword(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	original := adminbus.Admin{
		ID:           uuid.New(),
		FirstName:    name.MustParse("Testov"),
		LastName:     name.MustParse("Userov"),
		Login:        login.MustParse("test.user"),
		PasswordHash: "hashed:OldPass123456!",
	}
	store.data[original.ID] = original

	newPass := password.MustParse("NewPass456789!")
	updated, err := bus.Update(context.Background(), original.ID, adminbus.UpdateAdmin{
		Password: &newPass,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.PasswordHash != "hashed:NewPass456789!" {
		t.Errorf("PasswordHash = %q, want %q", updated.PasswordHash, "hashed:NewPass456789!")
	}
}

func TestUpdate_AdminNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	_, err := bus.Update(context.Background(), uuid.New(), adminbus.UpdateAdmin{})
	if !errors.Is(err, adminbus.ErrNotFound) {
		t.Fatalf("Update error = %v, want %v", err, adminbus.ErrNotFound)
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsAdmin(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	adm := adminbus.Admin{
		ID:        uuid.New(),
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("Admin"),
		Login:     login.MustParse("test.admin"),
	}
	store.data[adm.ID] = adm

	got, err := bus.QueryByID(context.Background(), adm.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != adm.ID {
		t.Errorf("ID = %v, want %v", got.ID, adm.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, adminbus.ErrNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, adminbus.ErrNotFound)
	}
}

// =============================================================================
// Authenticate tests

func TestAuthenticate_ValidCredentials(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	lgn := login.MustParse("auth.user")
	pass := password.MustParse("CorrectPass123456!")

	store.loginMap[lgn.String()] = adminbus.Admin{
		ID:           uuid.New(),
		Login:        lgn,
		PasswordHash: "hashed:CorrectPass123456!",
	}

	adm, err := bus.Authenticate(context.Background(), lgn, pass)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if adm.Login != lgn {
		t.Errorf("Login = %v, want %v", adm.Login, lgn)
	}
}

func TestAuthenticate_WrongPassword_ReturnsErrInvalidCredential(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	lgn := login.MustParse("auth.user")
	store.loginMap[lgn.String()] = adminbus.Admin{
		ID:           uuid.New(),
		Login:        lgn,
		PasswordHash: "hashed:CorrectPass123456!",
	}

	_, err := bus.Authenticate(context.Background(), lgn, password.MustParse("WrongPass123456!"))
	if !errors.Is(err, adminbus.ErrInvalidCredential) {
		t.Fatalf("Authenticate error = %v, want %v", err, adminbus.ErrInvalidCredential)
	}
}

func TestAuthenticate_LoginNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	_, err := bus.Authenticate(context.Background(), login.MustParse("no.user"), password.MustParse("Pass123456789!"))
	if err == nil {
		t.Fatal("expected error for unknown login, got nil")
	}
}

// =============================================================================
// QueryByLogin tests

func TestQueryByLogin_ReturnsAdmin(t *testing.T) {
	t.Parallel()

	store := newAdminStorerStub()
	bus := adminbus.NewBusiness(&hasherStub{}, store)

	lgn := login.MustParse("find.me")
	adm := adminbus.Admin{
		ID:    uuid.New(),
		Login: lgn,
	}
	store.loginMap[lgn.String()] = adm

	got, err := bus.QueryByLogin(context.Background(), lgn)
	if err != nil {
		t.Fatalf("QueryByLogin returned error: %v", err)
	}
	if got.Login != lgn {
		t.Errorf("Login = %v, want %v", got.Login, lgn)
	}
}

func TestQueryByLogin_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	_, err := bus.QueryByLogin(context.Background(), login.MustParse("ghost"))
	if !errors.Is(err, adminbus.ErrNotFound) {
		t.Fatalf("QueryByLogin error = %v, want %v", err, adminbus.ErrNotFound)
	}
}
