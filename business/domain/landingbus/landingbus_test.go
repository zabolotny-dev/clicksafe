package landingbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type landingStorerStub struct {
	data    map[uuid.UUID]landingbus.Landing
	saveErr error
}

func newLandingStorerStub() *landingStorerStub {
	return &landingStorerStub{data: make(map[uuid.UUID]landingbus.Landing)}
}

func (s *landingStorerStub) Save(_ context.Context, l landingbus.Landing) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[l.ID] = l
	return nil
}

func (s *landingStorerStub) Update(_ context.Context, l landingbus.Landing) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[l.ID] = l
	return nil
}

func (s *landingStorerStub) Delete(_ context.Context, l landingbus.Landing) error {
	delete(s.data, l.ID)
	return nil
}

func (s *landingStorerStub) QueryByID(_ context.Context, id uuid.UUID) (landingbus.Landing, error) {
	l, ok := s.data[id]
	if !ok {
		return landingbus.Landing{}, landingbus.ErrNotFound
	}
	return l, nil
}

func (s *landingStorerStub) Query(_ context.Context, _ landingbus.QueryFilter, _ order.By, _ page.Page) ([]landingbus.Landing, error) {
	var result []landingbus.Landing
	for _, l := range s.data {
		result = append(result, l)
	}
	return result, nil
}

func (s *landingStorerStub) Count(_ context.Context, _ landingbus.QueryFilter) (int, error) {
	return len(s.data), nil
}

type landingAttachmentQuerierStub struct {
	data map[uuid.UUID]attachmentbus.Attachment
}

func newLandingAttachmentQuerierStub() *landingAttachmentQuerierStub {
	return &landingAttachmentQuerierStub{data: make(map[uuid.UUID]attachmentbus.Attachment)}
}

func (s *landingAttachmentQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	a, ok := s.data[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return a, nil
}

// =============================================================================
// Save tests

func TestSave_WithoutHtmlBody_CreatesLanding(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	l, err := bus.Save(context.Background(), landingbus.NewLanding{
		Label: label.MustParse("Welcome page"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if l.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if l.HtmlBodyID.Valid {
		t.Error("HtmlBodyID should not be set")
	}
	if _, exists := store.data[l.ID]; !exists {
		t.Error("Landing was not persisted to storer")
	}
}

func TestSave_WithHTMLAttachment_Success(t *testing.T) {
	t.Parallel()

	aq := newLandingAttachmentQuerierStub()
	htmlID := uuid.New()
	aq.data[htmlID] = attachmentbus.Attachment{
		ID:   htmlID,
		Type: attachmentbus.Html,
	}

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, aq)

	l, err := bus.Save(context.Background(), landingbus.NewLanding{
		Label:      label.MustParse("Phishing page"),
		HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if !l.HtmlBodyID.Valid || l.HtmlBodyID.UUID != htmlID {
		t.Errorf("HtmlBodyID = %v, want %v", l.HtmlBodyID, htmlID)
	}
}

func TestSave_WithNonHTMLAttachment_ReturnsErrInvalidAttachment(t *testing.T) {
	t.Parallel()

	aq := newLandingAttachmentQuerierStub()
	pngID := uuid.New()
	aq.data[pngID] = attachmentbus.Attachment{
		ID:   pngID,
		Type: attachmentbus.Png,
	}

	bus := landingbus.NewBusiness(newLandingStorerStub(), aq)

	_, err := bus.Save(context.Background(), landingbus.NewLanding{
		Label:      label.MustParse("Bad landing"),
		HtmlBodyID: uuid.NullUUID{UUID: pngID, Valid: true},
	})
	if !errors.Is(err, landingbus.ErrInvalidAttachment) {
		t.Fatalf("Save error = %v, want %v", err, landingbus.ErrInvalidAttachment)
	}
}

func TestSave_AttachmentNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := landingbus.NewBusiness(newLandingStorerStub(), newLandingAttachmentQuerierStub())

	_, err := bus.Save(context.Background(), landingbus.NewLanding{
		Label:      label.MustParse("No attachment"),
		HtmlBodyID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
	})
	if err == nil {
		t.Fatal("expected error for missing attachment, got nil")
	}
}

func TestSave_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	store.saveErr = errors.New("db error")
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	_, err := bus.Save(context.Background(), landingbus.NewLanding{
		Label: label.MustParse("Page"),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	original := landingbus.Landing{
		ID:    uuid.New(),
		Label: label.MustParse("OldLabel"),
	}
	store.data[original.ID] = original

	newLabel := label.MustParse("NewLabel")
	updated, err := bus.Update(context.Background(), original, landingbus.UpdateLanding{
		Label: &newLabel,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
}

func TestUpdate_SetsHTMLBody(t *testing.T) {
	t.Parallel()

	aq := newLandingAttachmentQuerierStub()
	htmlID := uuid.New()
	aq.data[htmlID] = attachmentbus.Attachment{
		ID:   htmlID,
		Type: attachmentbus.Html,
	}

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, aq)

	original := landingbus.Landing{
		ID:    uuid.New(),
		Label: label.MustParse("Landing"),
	}
	store.data[original.ID] = original

	htmlBody := uuid.NullUUID{UUID: htmlID, Valid: true}
	updated, err := bus.Update(context.Background(), original, landingbus.UpdateLanding{
		HtmlBodyID: &htmlBody,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !updated.HtmlBodyID.Valid || updated.HtmlBodyID.UUID != htmlID {
		t.Errorf("HtmlBodyID = %v, want %v", updated.HtmlBodyID, htmlID)
	}
}

func TestUpdate_ClearsHTMLBody(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	existingID := uuid.New()
	original := landingbus.Landing{
		ID:         uuid.New(),
		Label:      label.MustParse("Landing"),
		HtmlBodyID: uuid.NullUUID{UUID: existingID, Valid: true},
	}
	store.data[original.ID] = original

	// Сброс HtmlBodyID через null-значение
	nullBody := uuid.NullUUID{Valid: false}
	updated, err := bus.Update(context.Background(), original, landingbus.UpdateLanding{
		HtmlBodyID: &nullBody,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.HtmlBodyID.Valid {
		t.Error("HtmlBodyID should be cleared")
	}
}

func TestUpdate_NonHTMLAttachment_ReturnsError(t *testing.T) {
	t.Parallel()

	aq := newLandingAttachmentQuerierStub()
	txtID := uuid.New()
	aq.data[txtID] = attachmentbus.Attachment{
		ID:   txtID,
		Type: attachmentbus.Txt,
	}

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, aq)

	original := landingbus.Landing{
		ID:    uuid.New(),
		Label: label.MustParse("Landing"),
	}
	store.data[original.ID] = original

	txtBody := uuid.NullUUID{UUID: txtID, Valid: true}
	_, err := bus.Update(context.Background(), original, landingbus.UpdateLanding{
		HtmlBodyID: &txtBody,
	})
	if !errors.Is(err, landingbus.ErrInvalidAttachment) {
		t.Fatalf("Update error = %v, want %v", err, landingbus.ErrInvalidAttachment)
	}
}

// =============================================================================
// Delete tests

func TestDelete_RemovesFromStorer(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	l := landingbus.Landing{
		ID:    uuid.New(),
		Label: label.MustParse("ToDelete"),
	}
	store.data[l.ID] = l

	if err := bus.Delete(context.Background(), l); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, exists := store.data[l.ID]; exists {
		t.Error("Landing still exists in storer after Delete")
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsLanding(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	l := landingbus.Landing{
		ID:    uuid.New(),
		Label: label.MustParse("Find me"),
	}
	store.data[l.ID] = l

	got, err := bus.QueryByID(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != l.ID {
		t.Errorf("ID = %v, want %v", got.ID, l.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := landingbus.NewBusiness(newLandingStorerStub(), newLandingAttachmentQuerierStub())

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, landingbus.ErrNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, landingbus.ErrNotFound)
	}
}

// =============================================================================
// Count tests

func TestCount_ReturnsCorrectNumber(t *testing.T) {
	t.Parallel()

	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	for i := 0; i < 4; i++ {
		store.data[uuid.New()] = landingbus.Landing{ID: uuid.New()}
	}

	count, err := bus.Count(context.Background(), landingbus.QueryFilter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 4 {
		t.Errorf("Count = %d, want 4", count)
	}
}
