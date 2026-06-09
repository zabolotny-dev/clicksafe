package departmentbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type departmentStorerStub struct {
	data        map[uuid.UUID]departmentbus.Department
	saveErr     error
	saveManyErr error
	saveManyN   int // количество вызовов SaveMany
}

func newDepartmentStorerStub() *departmentStorerStub {
	return &departmentStorerStub{data: make(map[uuid.UUID]departmentbus.Department)}
}

func (s *departmentStorerStub) Save(_ context.Context, d departmentbus.Department) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[d.ID] = d
	return nil
}

func (s *departmentStorerStub) SaveMany(_ context.Context, deps []departmentbus.Department) error {
	s.saveManyN++
	if s.saveManyErr != nil {
		return s.saveManyErr
	}
	for _, d := range deps {
		s.data[d.ID] = d
	}
	return nil
}

func (s *departmentStorerStub) Update(_ context.Context, d departmentbus.Department) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[d.ID] = d
	return nil
}

func (s *departmentStorerStub) Delete(_ context.Context, d departmentbus.Department) error {
	delete(s.data, d.ID)
	return nil
}

func (s *departmentStorerStub) QueryByID(_ context.Context, id uuid.UUID) (departmentbus.Department, error) {
	d, ok := s.data[id]
	if !ok {
		return departmentbus.Department{}, departmentbus.ErrNotFound
	}
	return d, nil
}

func (s *departmentStorerStub) Query(_ context.Context, _ departmentbus.QueryFilter, _ order.By, _ page.Page) ([]departmentbus.Department, error) {
	var result []departmentbus.Department
	for _, d := range s.data {
		result = append(result, d)
	}
	return result, nil
}

func (s *departmentStorerStub) Count(_ context.Context, _ departmentbus.QueryFilter) (int, error) {
	return len(s.data), nil
}

// =============================================================================
// Save tests

func TestSave_AssignsNonZeroUUID(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	dep, err := bus.Save(context.Background(), departmentbus.NewDepartment{
		Label: label.MustParse("Engineering"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if dep.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if dep.Label != label.MustParse("Engineering") {
		t.Errorf("Label = %v, want Engineering", dep.Label)
	}
	if _, exists := store.data[dep.ID]; !exists {
		t.Error("Department was not persisted to storer")
	}
}

func TestSave_SetsAttributes(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	attrs := map[string]string{"Slack": "#eng"}
	dep, err := bus.Save(context.Background(), departmentbus.NewDepartment{
		Label:      label.MustParse("Engineering"),
		Attributes: attrs,
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if dep.Attributes["Slack"] != "#eng" {
		t.Errorf("Attributes[Slack] = %q, want %q", dep.Attributes["Slack"], "#eng")
	}
}

func TestSave_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	store.saveErr = errors.New("db error")
	bus := departmentbus.NewBusiness(store)

	_, err := bus.Save(context.Background(), departmentbus.NewDepartment{
		Label: label.MustParse("Engineering"),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// SaveMany tests

func TestSaveMany_Empty_DoesNotCallStorer(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	err := bus.SaveMany(context.Background(), []departmentbus.NewDepartment{})
	if err != nil {
		t.Fatalf("SaveMany returned error: %v", err)
	}
	if store.saveManyN != 0 {
		t.Errorf("SaveMany called storer %d times, want 0", store.saveManyN)
	}
}

func TestSaveMany_MultipleItems_PersistsAll(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	err := bus.SaveMany(context.Background(), []departmentbus.NewDepartment{
		{Label: label.MustParse("HR dept")},
		{Label: label.MustParse("Finance dept")},
		{Label: label.MustParse("Engineering")},
	})
	if err != nil {
		t.Fatalf("SaveMany returned error: %v", err)
	}
	if len(store.data) != 3 {
		t.Errorf("storer has %d departments, want 3", len(store.data))
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	original := departmentbus.Department{
		ID:    uuid.New(),
		Label: label.MustParse("OldLabel"),
	}
	store.data[original.ID] = original

	newLabel := label.MustParse("NewLabel")
	updated, err := bus.Update(context.Background(), original, departmentbus.UpdateDepartment{
		Label: &newLabel,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
	// Убеждаемся что в стабе тоже обновилось
	if store.data[original.ID].Label != newLabel {
		t.Error("Storer not updated with new label")
	}
}

func TestUpdate_ChangesAttributes(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	original := departmentbus.Department{
		ID:         uuid.New(),
		Label:      label.MustParse("HR dept"),
		Attributes: map[string]string{"Chat": "hr-chat"},
	}
	store.data[original.ID] = original

	newAttrs := map[string]string{"Chat": "new-hr-chat", "Email": "hr@example.com"}
	updated, err := bus.Update(context.Background(), original, departmentbus.UpdateDepartment{
		Attributes: &newAttrs,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Attributes["Email"] != "hr@example.com" {
		t.Errorf("Attributes[Email] = %q, want %q", updated.Attributes["Email"], "hr@example.com")
	}
}

func TestUpdate_NoChanges_ReturnsOriginal(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	original := departmentbus.Department{
		ID:    uuid.New(),
		Label: label.MustParse("Finance dept"),
	}
	store.data[original.ID] = original

	updated, err := bus.Update(context.Background(), original, departmentbus.UpdateDepartment{})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != original.Label {
		t.Errorf("Label changed unexpectedly: got %v, want %v", updated.Label, original.Label)
	}
}

// =============================================================================
// Delete tests

func TestDelete_RemovesFromStorer(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	dep := departmentbus.Department{
		ID:    uuid.New(),
		Label: label.MustParse("ToDelete"),
	}
	store.data[dep.ID] = dep

	if err := bus.Delete(context.Background(), dep); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, exists := store.data[dep.ID]; exists {
		t.Error("Department still exists in storer after Delete")
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsCorrectDepartment(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	dep := departmentbus.Department{
		ID:    uuid.New(),
		Label: label.MustParse("Legal"),
	}
	store.data[dep.ID] = dep

	got, err := bus.QueryByID(context.Background(), dep.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != dep.ID {
		t.Errorf("ID = %v, want %v", got.ID, dep.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := departmentbus.NewBusiness(newDepartmentStorerStub())

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, departmentbus.ErrNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, departmentbus.ErrNotFound)
	}
}

// =============================================================================
// Count tests

func TestCount_ReturnsCorrectNumber(t *testing.T) {
	t.Parallel()

	store := newDepartmentStorerStub()
	bus := departmentbus.NewBusiness(store)

	for i := 0; i < 3; i++ {
		store.data[uuid.New()] = departmentbus.Department{ID: uuid.New()}
	}

	count, err := bus.Count(context.Background(), departmentbus.QueryFilter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}
