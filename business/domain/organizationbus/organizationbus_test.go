package organizationbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
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

func Test_Organization(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testSave(), "save")
	unittest.Run(t, testGet(), "get")
	unittest.Run(t, testUpdate(), "update")
}

// =============================================================================

func testSave() []unittest.Table {
	storeOK := newOrgStorerStub()
	busOK := organizationbus.NewBusiness(storeOK, newOrgAttachmentQuerierStub())

	aqImage := newOrgAttachmentQuerierStub()
	imgID := uuid.New()
	aqImage.data[imgID] = attachmentbus.Attachment{ID: imgID, Type: attachmentbus.Png}
	storeImg := newOrgStorerStub()
	busImg := organizationbus.NewBusiness(storeImg, aqImage)

	aqNonImage := newOrgAttachmentQuerierStub()
	htmlID := uuid.New()
	aqNonImage.data[htmlID] = attachmentbus.Attachment{ID: htmlID, Type: attachmentbus.Html}
	busNonImage := organizationbus.NewBusiness(newOrgStorerStub(), aqNonImage)

	busAttNotFound := organizationbus.NewBusiness(newOrgStorerStub(), newOrgAttachmentQuerierStub())

	storeDBErr := newOrgStorerStub()
	storeDBErr.saveErr = errors.New("db error")
	busDBErr := organizationbus.NewBusiness(storeDBErr, newOrgAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "without-attachment-uses-global-id",
			ExpResp: organizationbus.GlobalID,
			ExcFunc: func(ctx context.Context) any {
				org, err := busOK.Save(ctx, organizationbus.NewOrganization{
					Label: label.MustParse("ClickSafe"),
				})
				if err != nil {
					return err
				}
				return org.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "with-image-attachment-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				org, err := busImg.Save(ctx, organizationbus.NewOrganization{
					Label:        label.MustParse("ClickSafe"),
					AttachmentID: uuid.NullUUID{UUID: imgID, Valid: true},
				})
				if err != nil {
					return err
				}
				return org.AttachmentID.Valid && org.AttachmentID.UUID == imgID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "with-non-image-attachment-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNonImage.Save(ctx, organizationbus.NewOrganization{
					Label:        label.MustParse("ClickSafe"),
					AttachmentID: uuid.NullUUID{UUID: htmlID, Valid: true},
				})
				return errors.Is(err, organizationbus.ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busAttNotFound.Save(ctx, organizationbus.NewOrganization{
					Label:        label.MustParse("ClickSafe"),
					AttachmentID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "with-attributes-persists",
			ExpResp: "support@clicksafe.test",
			ExcFunc: func(ctx context.Context) any {
				org, err := busOK.Save(ctx, organizationbus.NewOrganization{
					Label:      label.MustParse("ClickSafe org"),
					Attributes: map[string]string{"SupportEmail": "support@clicksafe.test"},
				})
				if err != nil {
					return err
				}
				return org.Attributes["SupportEmail"]
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDBErr.Save(ctx, organizationbus.NewOrganization{
					Label: label.MustParse("ClickSafe"),
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testGet() []unittest.Table {
	orgLabel := label.MustParse("ClickSafe")

	storeFound := newOrgStorerStub()
	storeFound.data[organizationbus.GlobalID] = organizationbus.Organization{
		ID:    organizationbus.GlobalID,
		Label: orgLabel,
	}
	busFound := organizationbus.NewBusiness(storeFound, newOrgAttachmentQuerierStub())

	busNotFound := organizationbus.NewBusiness(newOrgStorerStub(), newOrgAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-organization",
			ExpResp: organizationbus.GlobalID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busFound.Get(ctx)
				if err != nil {
					return err
				}
				return got.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNotFound.Get(ctx)
				return errors.Is(err, organizationbus.ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testUpdate() []unittest.Table {
	original := organizationbus.Organization{
		ID:         organizationbus.GlobalID,
		Label:      label.MustParse("OldLabel"),
		Attributes: map[string]string{"OldKey": "OldVal"},
	}

	storeLabel := newOrgStorerStub()
	storeLabel.data[original.ID] = original
	busLabel := organizationbus.NewBusiness(storeLabel, newOrgAttachmentQuerierStub())

	storeAttrs := newOrgStorerStub()
	storeAttrs.data[original.ID] = original
	busAttrs := organizationbus.NewBusiness(storeAttrs, newOrgAttachmentQuerierStub())

	aqImage := newOrgAttachmentQuerierStub()
	imgID := uuid.New()
	aqImage.data[imgID] = attachmentbus.Attachment{ID: imgID, Type: attachmentbus.Jpeg}
	storeImg := newOrgStorerStub()
	storeImg.data[original.ID] = original
	busImg := organizationbus.NewBusiness(storeImg, aqImage)

	aqNonImage := newOrgAttachmentQuerierStub()
	htmlID2 := uuid.New()
	aqNonImage.data[htmlID2] = attachmentbus.Attachment{ID: htmlID2, Type: attachmentbus.Html}
	storeNonImg := newOrgStorerStub()
	storeNonImg.data[original.ID] = original
	busNonImg := organizationbus.NewBusiness(storeNonImg, aqNonImage)

	newLabel := label.MustParse("NewLabel")
	newAttrs := map[string]string{"NewKey": "NewVal"}
	imgAttID := uuid.NullUUID{UUID: imgID, Valid: true}
	nonImgAttID := uuid.NullUUID{UUID: htmlID2, Valid: true}

	return []unittest.Table{
		{
			Name:    "changes-label",
			ExpResp: newLabel,
			ExcFunc: func(ctx context.Context) any {
				updated, err := busLabel.Update(ctx, original, organizationbus.UpdateOrganization{Label: &newLabel})
				if err != nil {
					return err
				}
				return updated.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "changes-attributes",
			ExpResp: "NewVal",
			ExcFunc: func(ctx context.Context) any {
				updated, err := busAttrs.Update(ctx, original, organizationbus.UpdateOrganization{Attributes: &newAttrs})
				if err != nil {
					return err
				}
				return updated.Attributes["NewKey"]
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "sets-image-attachment",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				updated, err := busImg.Update(ctx, original, organizationbus.UpdateOrganization{AttachmentID: &imgAttID})
				if err != nil {
					return err
				}
				return updated.AttachmentID.Valid && updated.AttachmentID.UUID == imgID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "non-image-attachment-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNonImg.Update(ctx, original, organizationbus.UpdateOrganization{AttachmentID: &nonImgAttID})
				return errors.Is(err, organizationbus.ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
