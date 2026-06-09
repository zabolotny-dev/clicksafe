package employeebus_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

// =============================================================================
// Stubs

type employeeStorerStub struct {
	data        map[uuid.UUID]employeebus.Employee
	saveErr     error
	saveManyErr error
	saveManyN   int
}

func newEmployeeStorerStub() *employeeStorerStub {
	return &employeeStorerStub{data: make(map[uuid.UUID]employeebus.Employee)}
}

func (s *employeeStorerStub) Save(_ context.Context, e employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[e.ID] = e
	return nil
}

func (s *employeeStorerStub) SaveMany(_ context.Context, emps []employeebus.Employee) error {
	s.saveManyN++
	if s.saveManyErr != nil {
		return s.saveManyErr
	}
	for _, e := range emps {
		s.data[e.ID] = e
	}
	return nil
}

func (s *employeeStorerStub) QueryByID(_ context.Context, id uuid.UUID) (employeebus.Employee, error) {
	e, ok := s.data[id]
	if !ok {
		return employeebus.Employee{}, employeebus.ErrNotFound
	}
	return e, nil
}

func (s *employeeStorerStub) Query(_ context.Context, _ employeebus.QueryFilter, _ order.By, _ page.Page) ([]employeebus.Employee, error) {
	var result []employeebus.Employee
	for _, e := range s.data {
		result = append(result, e)
	}
	return result, nil
}

func (s *employeeStorerStub) Update(_ context.Context, e employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[e.ID] = e
	return nil
}

func (s *employeeStorerStub) Delete(_ context.Context, e employeebus.Employee) error {
	delete(s.data, e.ID)
	return nil
}

func (s *employeeStorerStub) Count(_ context.Context, _ employeebus.QueryFilter) (int, error) {
	return len(s.data), nil
}

// mustParsePhone — вспомогательная функция для тестов
func mustParsePhone(v string) phone.Null {
	p, err := phone.ParseNull(v)
	if err != nil {
		panic(err)
	}
	return p
}

// =============================================================================
// Save tests

func TestSave_AssignsNonZeroUUID(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	emp, err := bus.Save(context.Background(), employeebus.NewEmployee{
		FirstName: name.MustParse("Ivan"),
		LastName:  name.MustParse("Petrov"),
		Email:     mail.Address{Address: "ivan@example.com"},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if emp.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if _, exists := store.data[emp.ID]; !exists {
		t.Error("Employee was not persisted to storer")
	}
}

func TestSave_WithDepartmentID_Persists(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	deptID := uuid.New()
	emp, err := bus.Save(context.Background(), employeebus.NewEmployee{
		DepartmentID: &deptID,
		FirstName:    name.MustParse("Petr"),
		LastName:     name.MustParse("Ivanov"),
		Email:        mail.Address{Address: "petr@example.com"},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if emp.DepartmentID == nil || *emp.DepartmentID != deptID {
		t.Errorf("DepartmentID = %v, want %v", emp.DepartmentID, deptID)
	}
}

func TestSave_DepartmentNotFound_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	store.saveErr = employeebus.ErrDepartmentNotFound
	bus := employeebus.NewBusiness(store)

	deptID := uuid.New()
	_, err := bus.Save(context.Background(), employeebus.NewEmployee{
		DepartmentID: &deptID,
		FirstName:    name.MustParse("Test"),
		LastName:     name.MustParse("User"),
		Email:        mail.Address{Address: "test@example.com"},
	})
	if !errors.Is(err, employeebus.ErrDepartmentNotFound) {
		t.Fatalf("Save error = %v, want %v", err, employeebus.ErrDepartmentNotFound)
	}
}

func TestSave_WithPhone_Persists(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	ph := mustParsePhone("+79991234567")
	emp, err := bus.Save(context.Background(), employeebus.NewEmployee{
		FirstName: name.MustParse("Anna"),
		LastName:  name.MustParse("Smirnova"),
		Email:     mail.Address{Address: "anna@example.com"},
		Phone:     ph,
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if emp.Phone.String() != ph.String() {
		t.Errorf("Phone = %v, want %v", emp.Phone, ph)
	}
}

func TestSave_WithAttributes_Persists(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	attrs := map[string]string{"Position": "Developer", "Level": "Senior"}
	emp, err := bus.Save(context.Background(), employeebus.NewEmployee{
		FirstName:  name.MustParse("Dev"),
		LastName:   name.MustParse("User"),
		Email:      mail.Address{Address: "dev@example.com"},
		Attributes: attrs,
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if emp.Attributes["Position"] != "Developer" {
		t.Errorf("Attributes[Position] = %q, want Developer", emp.Attributes["Position"])
	}
}

// =============================================================================
// SaveMany tests

func TestSaveMany_Empty_DoesNotCallStorer(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	err := bus.SaveMany(context.Background(), []employeebus.NewEmployee{})
	if err != nil {
		t.Fatalf("SaveMany returned error: %v", err)
	}
	if store.saveManyN != 0 {
		t.Errorf("SaveMany called storer %d times, want 0", store.saveManyN)
	}
}

func TestSaveMany_MultipleItems_PersistsAll(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	err := bus.SaveMany(context.Background(), []employeebus.NewEmployee{
		{FirstName: name.MustParse("Alice"), LastName: name.MustParse("Bo"), Email: mail.Address{Address: "alice@test.com"}},
		{FirstName: name.MustParse("Carol"), LastName: name.MustParse("Do"), Email: mail.Address{Address: "carol@test.com"}},
	})
	if err != nil {
		t.Fatalf("SaveMany returned error: %v", err)
	}
	if len(store.data) != 2 {
		t.Errorf("storer has %d employees, want 2", len(store.data))
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesFirstName(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	original := employeebus.Employee{
		ID:        uuid.New(),
		FirstName: name.MustParse("OldName"),
		LastName:  name.MustParse("LastName"),
		Email:     mail.Address{Address: "old@example.com"},
	}
	store.data[original.ID] = original

	newFirst := name.MustParse("NewName")
	updated, err := bus.Update(context.Background(), original, employeebus.UpdateEmployee{
		FirstName: &newFirst,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.FirstName != newFirst {
		t.Errorf("FirstName = %v, want %v", updated.FirstName, newFirst)
	}
	if updated.LastName != original.LastName {
		t.Errorf("LastName changed unexpectedly: got %v, want %v", updated.LastName, original.LastName)
	}
}

func TestUpdate_ChangesEmail(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	original := employeebus.Employee{
		ID:        uuid.New(),
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("User"),
		Email:     mail.Address{Address: "old@example.com"},
	}
	store.data[original.ID] = original

	newEmail := mail.Address{Address: "new@example.com"}
	updated, err := bus.Update(context.Background(), original, employeebus.UpdateEmployee{
		Email: &newEmail,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Email.Address != newEmail.Address {
		t.Errorf("Email = %v, want %v", updated.Email, newEmail)
	}
}

func TestUpdate_ChangesDepartmentID(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	oldDeptID := uuid.New()
	original := employeebus.Employee{
		ID:           uuid.New(),
		DepartmentID: &oldDeptID,
		FirstName:    name.MustParse("Test"),
		LastName:     name.MustParse("User"),
		Email:        mail.Address{Address: "test@example.com"},
	}
	store.data[original.ID] = original

	newDeptID := uuid.New()
	updated, err := bus.Update(context.Background(), original, employeebus.UpdateEmployee{
		DepartmentID: &newDeptID,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.DepartmentID == nil || *updated.DepartmentID != newDeptID {
		t.Errorf("DepartmentID = %v, want %v", updated.DepartmentID, newDeptID)
	}
}

func TestUpdate_DepartmentNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	store.saveErr = employeebus.ErrDepartmentNotFound
	bus := employeebus.NewBusiness(store)

	original := employeebus.Employee{
		ID:        uuid.New(),
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("User"),
		Email:     mail.Address{Address: "test@example.com"},
	}
	store.data[original.ID] = original

	newDeptID := uuid.New()
	_, err := bus.Update(context.Background(), original, employeebus.UpdateEmployee{
		DepartmentID: &newDeptID,
	})
	if !errors.Is(err, employeebus.ErrDepartmentNotFound) {
		t.Fatalf("Update error = %v, want %v", err, employeebus.ErrDepartmentNotFound)
	}
}

func TestUpdate_ChangesAttributes(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	original := employeebus.Employee{
		ID:         uuid.New(),
		FirstName:  name.MustParse("Test"),
		LastName:   name.MustParse("User"),
		Email:      mail.Address{Address: "test@example.com"},
		Attributes: map[string]string{"Level": "Junior"},
	}
	store.data[original.ID] = original

	newAttrs := map[string]string{"Level": "Senior", "Team": "Backend"}
	updated, err := bus.Update(context.Background(), original, employeebus.UpdateEmployee{
		Attributes: &newAttrs,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Attributes["Level"] != "Senior" {
		t.Errorf("Attributes[Level] = %q, want Senior", updated.Attributes["Level"])
	}
}

// =============================================================================
// Delete tests

func TestDelete_RemovesFromStorer(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	emp := employeebus.Employee{
		ID:        uuid.New(),
		FirstName: name.MustParse("ToDelete"),
		LastName:  name.MustParse("User"),
		Email:     mail.Address{Address: "delete@example.com"},
	}
	store.data[emp.ID] = emp

	if err := bus.Delete(context.Background(), emp); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, exists := store.data[emp.ID]; exists {
		t.Error("Employee still exists in storer after Delete")
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsEmployee(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	emp := employeebus.Employee{
		ID:        uuid.New(),
		FirstName: name.MustParse("Find"),
		LastName:  name.MustParse("Me"),
		Email:     mail.Address{Address: "find@example.com"},
	}
	store.data[emp.ID] = emp

	got, err := bus.QueryByID(context.Background(), emp.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != emp.ID {
		t.Errorf("ID = %v, want %v", got.ID, emp.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := employeebus.NewBusiness(newEmployeeStorerStub())

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, employeebus.ErrNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, employeebus.ErrNotFound)
	}
}

// =============================================================================
// Count tests

func TestCount_ReturnsCorrectNumber(t *testing.T) {
	t.Parallel()

	store := newEmployeeStorerStub()
	bus := employeebus.NewBusiness(store)

	for i := 0; i < 5; i++ {
		store.data[uuid.New()] = employeebus.Employee{ID: uuid.New()}
	}

	count, err := bus.Count(context.Background(), employeebus.QueryFilter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 5 {
		t.Errorf("Count = %d, want 5", count)
	}
}
