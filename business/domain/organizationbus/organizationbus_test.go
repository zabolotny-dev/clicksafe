package organizationbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type orgStorerStub struct {
	data    map[uuid.UUID]organizationbus.Organization
	saveErr error
}

func newOrgStorerStub() *orgStorerStub {
	return &orgStorerStub{data: make(map[uuid.UUID]organizationbus.Organization)}
}

func (s *orgStorerStub) Save(_ context.Context, org organizationbus.Organization) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[org.ID] = org
	return nil
}

func (s *orgStorerStub) QueryByID(_ context.Context, id uuid.UUID) (organizationbus.Organization, error) {
	org, ok := s.data[id]
	if !ok {
		return organizationbus.Organization{}, organizationbus.ErrNotFound
	}
	return org, nil
}

func (s *orgStorerStub) Update(_ context.Context, org organizationbus.Organization) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[org.ID] = org
	return nil
}

type orgAttachmentQuerierStub struct {
	data map[uuid.UUID]attachmentbus.Attachment
}

func newOrgAttachmentQuerierStub() *orgAttachmentQuerierStub {
	return &orgAttachmentQuerierStub{data: make(map[uuid.UUID]attachmentbus.Attachment)}
}

func (s *orgAttachmentQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	a, ok := s.data[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return a, nil
}

// =============================================================================
// Save tests

func TestSave_WithoutAttachment_UsesGlobalID(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	org, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label: label.MustParse("ClickSafe"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if org.ID != organizationbus.GlobalID {
		t.Errorf("ID = %v, want GlobalID = %v", org.ID, organizationbus.GlobalID)
	}
	if _, exists := store.data[org.ID]; !exists {
		t.Error("Organization was not persisted to storer")
	}
}

func TestSave_WithImageAttachment_Success(t *testing.T) {
	t.Parallel()

	aq := newOrgAttachmentQuerierStub()
	imgID := uuid.New()
	aq.data[imgID] = attachmentbus.Attachment{
		ID:   imgID,
		Type: attachmentbus.Png,
	}

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, aq)

	org, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label:        label.MustParse("ClickSafe"),
		AttachmentID: uuid.NullUUID{UUID: imgID, Valid: true},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if !org.AttachmentID.Valid || org.AttachmentID.UUID != imgID {
		t.Errorf("AttachmentID = %v, want %v", org.AttachmentID, imgID)
	}
}

func TestSave_WithNonImageAttachment_ReturnsErrInvalidAttachment(t *testing.T) {
	t.Parallel()

	aq := newOrgAttachmentQuerierStub()
	htmlID := uuid.New()
	aq.data[htmlID] = attachmentbus.Attachment{
		ID:   htmlID,
		Type: attachmentbus.Html,
	}

	bus := organizationbus.NewBusiness(newOrgStorerStub(), aq)

	_, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label:        label.MustParse("ClickSafe"),
		AttachmentID: uuid.NullUUID{UUID: htmlID, Valid: true},
	})
	if !errors.Is(err, organizationbus.ErrInvalidAttachment) {
		t.Fatalf("Save error = %v, want %v", err, organizationbus.ErrInvalidAttachment)
	}
}

func TestSave_AttachmentNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := organizationbus.NewBusiness(newOrgStorerStub(), newOrgAttachmentQuerierStub())

	_, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label:        label.MustParse("ClickSafe"),
		AttachmentID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
	})
	if err == nil {
		t.Fatal("expected error for missing attachment, got nil")
	}
}

func TestSave_WithAttributes_Persists(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	attrs := map[string]string{"SupportEmail": "support@clicksafe.test"}
	org, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label:      label.MustParse("ClickSafe"),
		Attributes: attrs,
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if org.Attributes["SupportEmail"] != "support@clicksafe.test" {
		t.Errorf("Attributes[SupportEmail] = %q, want support@clicksafe.test", org.Attributes["SupportEmail"])
	}
}

func TestSave_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	store.saveErr = errors.New("db error")
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	_, err := bus.Save(context.Background(), organizationbus.NewOrganization{
		Label: label.MustParse("ClickSafe"),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// Get tests

func TestGet_ReturnsOrganization(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	org := organizationbus.Organization{
		ID:    organizationbus.GlobalID,
		Label: label.MustParse("ClickSafe"),
	}
	store.data[org.ID] = org

	got, err := bus.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != org.ID {
		t.Errorf("ID = %v, want %v", got.ID, org.ID)
	}
	if got.Label != org.Label {
		t.Errorf("Label = %v, want %v", got.Label, org.Label)
	}
}

func TestGet_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := organizationbus.NewBusiness(newOrgStorerStub(), newOrgAttachmentQuerierStub())

	_, err := bus.Get(context.Background())
	if !errors.Is(err, organizationbus.ErrNotFound) {
		t.Fatalf("Get error = %v, want %v", err, organizationbus.ErrNotFound)
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	original := organizationbus.Organization{
		ID:    organizationbus.GlobalID,
		Label: label.MustParse("OldLabel"),
	}
	store.data[original.ID] = original

	newLabel := label.MustParse("NewLabel")
	updated, err := bus.Update(context.Background(), original, organizationbus.UpdateOrganization{
		Label: &newLabel,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
}

func TestUpdate_ChangesAttributes(t *testing.T) {
	t.Parallel()

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, newOrgAttachmentQuerierStub())

	original := organizationbus.Organization{
		ID:         organizationbus.GlobalID,
		Label:      label.MustParse("ClickSafe"),
		Attributes: map[string]string{"OldKey": "OldVal"},
	}
	store.data[original.ID] = original

	newAttrs := map[string]string{"NewKey": "NewVal"}
	updated, err := bus.Update(context.Background(), original, organizationbus.UpdateOrganization{
		Attributes: &newAttrs,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Attributes["NewKey"] != "NewVal" {
		t.Errorf("Attributes[NewKey] = %q, want NewVal", updated.Attributes["NewKey"])
	}
}

func TestUpdate_SetsImageAttachment(t *testing.T) {
	t.Parallel()

	aq := newOrgAttachmentQuerierStub()
	imgID := uuid.New()
	aq.data[imgID] = attachmentbus.Attachment{
		ID:   imgID,
		Type: attachmentbus.Jpeg,
	}

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, aq)

	original := organizationbus.Organization{
		ID:    organizationbus.GlobalID,
		Label: label.MustParse("ClickSafe"),
	}
	store.data[original.ID] = original

	att := uuid.NullUUID{UUID: imgID, Valid: true}
	updated, err := bus.Update(context.Background(), original, organizationbus.UpdateOrganization{
		AttachmentID: &att,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !updated.AttachmentID.Valid || updated.AttachmentID.UUID != imgID {
		t.Errorf("AttachmentID = %v, want %v", updated.AttachmentID, imgID)
	}
}

func TestUpdate_WithNonImageAttachment_ReturnsError(t *testing.T) {
	t.Parallel()

	aq := newOrgAttachmentQuerierStub()
	htmlID := uuid.New()
	aq.data[htmlID] = attachmentbus.Attachment{
		ID:   htmlID,
		Type: attachmentbus.Html,
	}

	store := newOrgStorerStub()
	bus := organizationbus.NewBusiness(store, aq)

	original := organizationbus.Organization{
		ID:    organizationbus.GlobalID,
		Label: label.MustParse("ClickSafe"),
	}
	store.data[original.ID] = original

	att := uuid.NullUUID{UUID: htmlID, Valid: true}
	_, err := bus.Update(context.Background(), original, organizationbus.UpdateOrganization{
		AttachmentID: &att,
	})
	if !errors.Is(err, organizationbus.ErrInvalidAttachment) {
		t.Fatalf("Update error = %v, want %v", err, organizationbus.ErrInvalidAttachment)
	}
}
