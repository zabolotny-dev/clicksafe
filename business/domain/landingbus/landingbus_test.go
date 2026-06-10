package landingbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type landingStorerStub struct {
	data      map[uuid.UUID]landingbus.Landing
	saveErr   error
	deleteErr error
	queryErr  error
	countErr  error
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
	if s.deleteErr != nil {
		return s.deleteErr
	}
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
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	var result []landingbus.Landing
	for _, l := range s.data {
		result = append(result, l)
	}
	return result, nil
}

func (s *landingStorerStub) Count(_ context.Context, _ landingbus.QueryFilter) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
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

func Test_Landing(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testSave(), "save")
	unittest.Run(t, testUpdate(), "update")
	unittest.Run(t, testDelete(), "delete")
	unittest.Run(t, testQuery(), "query")
	unittest.Run(t, testQueryByID(), "querybyid")
	unittest.Run(t, testCount(), "count")
	unittest.Run(t, testLandingSeedHelper(), "seed-helper")
}

func testLandingSeedHelper() []unittest.Table {
	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "seed-creates-n-landings",
			ExpResp: 3,
			ExcFunc: func(ctx context.Context) any {
				landings, err := landingbus.TestSeedLandings(ctx, 3, bus)
				if err != nil {
					return err
				}
				return len(landings)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testQuery() []unittest.Table {
	store := newLandingStorerStub()
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	for i := range 2 {
		store.data[uuid.New()] = landingbus.Landing{
			ID:    uuid.New(),
			Label: label.MustParse(fmt.Sprintf("Landing %d of test", i+1)),
		}
	}

	storeErr := newLandingStorerStub()
	storeErr.queryErr = errors.New("db error")
	busErr := landingbus.NewBusiness(storeErr, newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-all",
			ExpResp: 2,
			ExcFunc: func(ctx context.Context) any {
				result, err := bus.Query(ctx, landingbus.QueryFilter{}, landingbus.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.Query(ctx, landingbus.QueryFilter{}, landingbus.DefaultOrderBy, page.MustParse("1", "10"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

// =============================================================================

func testSave() []unittest.Table {
	storeOK := newLandingStorerStub()
	busOK := landingbus.NewBusiness(storeOK, newLandingAttachmentQuerierStub())

	htmlID := uuid.New()
	aqHTML := newLandingAttachmentQuerierStub()
	aqHTML.data[htmlID] = attachmentbus.Attachment{ID: htmlID, Type: attachmentbus.Html}
	storeHTML := newLandingStorerStub()
	busHTML := landingbus.NewBusiness(storeHTML, aqHTML)

	pngID := uuid.New()
	aqPNG := newLandingAttachmentQuerierStub()
	aqPNG.data[pngID] = attachmentbus.Attachment{ID: pngID, Type: attachmentbus.Png}
	busNonHTML := landingbus.NewBusiness(newLandingStorerStub(), aqPNG)

	busAttNotFound := landingbus.NewBusiness(newLandingStorerStub(), newLandingAttachmentQuerierStub())

	storeDBErr := newLandingStorerStub()
	storeDBErr.saveErr = errors.New("db error")
	busDBErr := landingbus.NewBusiness(storeDBErr, newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "without-html-body-creates-landing",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				l, err := busOK.Save(ctx, landingbus.NewLanding{Label: label.MustParse("Welcome page")})
				if err != nil {
					return err
				}
				return l.ID != (uuid.UUID{}) && !l.HtmlBodyID.Valid && storeOK.data[l.ID].Label == l.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "with-html-attachment-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				l, err := busHTML.Save(ctx, landingbus.NewLanding{
					Label:      label.MustParse("Phishing page"),
					HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
				})
				if err != nil {
					return err
				}
				return l.HtmlBodyID.Valid && l.HtmlBodyID.UUID == htmlID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "non-html-attachment-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNonHTML.Save(ctx, landingbus.NewLanding{
					Label:      label.MustParse("Bad landing"),
					HtmlBodyID: uuid.NullUUID{UUID: pngID, Valid: true},
				})
				return errors.Is(err, landingbus.ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busAttNotFound.Save(ctx, landingbus.NewLanding{
					Label:      label.MustParse("No attachment"),
					HtmlBodyID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDBErr.Save(ctx, landingbus.NewLanding{Label: label.MustParse("Page")})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testUpdate() []unittest.Table {
	original := landingbus.Landing{ID: uuid.New(), Label: label.MustParse("OldLabel")}

	storeLabel := newLandingStorerStub()
	storeLabel.data[original.ID] = original
	busLabel := landingbus.NewBusiness(storeLabel, newLandingAttachmentQuerierStub())

	htmlID2 := uuid.New()
	aqHTML2 := newLandingAttachmentQuerierStub()
	aqHTML2.data[htmlID2] = attachmentbus.Attachment{ID: htmlID2, Type: attachmentbus.Html}
	storeHTMLUp := newLandingStorerStub()
	storeHTMLUp.data[original.ID] = original
	busHTMLUp := landingbus.NewBusiness(storeHTMLUp, aqHTML2)

	origWithHTML := landingbus.Landing{
		ID:         uuid.New(),
		Label:      label.MustParse("WithHTML"),
		HtmlBodyID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
	}
	storeClear := newLandingStorerStub()
	storeClear.data[origWithHTML.ID] = origWithHTML
	busClear := landingbus.NewBusiness(storeClear, newLandingAttachmentQuerierStub())

	txtID := uuid.New()
	aqTXT := newLandingAttachmentQuerierStub()
	aqTXT.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt}
	storeTXT := newLandingStorerStub()
	storeTXT.data[original.ID] = original
	busTXT := landingbus.NewBusiness(storeTXT, aqTXT)

	newLabel := label.MustParse("NewLabel")
	htmlBody := uuid.NullUUID{UUID: htmlID2, Valid: true}
	nullBody := uuid.NullUUID{Valid: false}
	txtBody := uuid.NullUUID{UUID: txtID, Valid: true}

	return []unittest.Table{
		{
			Name:    "changes-label",
			ExpResp: newLabel,
			ExcFunc: func(ctx context.Context) any {
				updated, err := busLabel.Update(ctx, original, landingbus.UpdateLanding{Label: &newLabel})
				if err != nil {
					return err
				}
				return updated.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "sets-html-body",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				updated, err := busHTMLUp.Update(ctx, original, landingbus.UpdateLanding{HtmlBodyID: &htmlBody})
				if err != nil {
					return err
				}
				return updated.HtmlBodyID.Valid && updated.HtmlBodyID.UUID == htmlID2
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "clears-html-body",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				updated, err := busClear.Update(ctx, origWithHTML, landingbus.UpdateLanding{HtmlBodyID: &nullBody})
				if err != nil {
					return err
				}
				return updated.HtmlBodyID.Valid
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "non-html-attachment-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busTXT.Update(ctx, original, landingbus.UpdateLanding{HtmlBodyID: &txtBody})
				return errors.Is(err, landingbus.ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				unknownID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
				_, err := busLabel.Update(ctx, original, landingbus.UpdateLanding{HtmlBodyID: &unknownID})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-update-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				storeUpErr := newLandingStorerStub()
				storeUpErr.data[original.ID] = original
				storeUpErr.saveErr = errors.New("db error")
				busUpErr := landingbus.NewBusiness(storeUpErr, newLandingAttachmentQuerierStub())
				newLbl := label.MustParse("AnyLabel")
				_, err := busUpErr.Update(ctx, original, landingbus.UpdateLanding{Label: &newLbl})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testDelete() []unittest.Table {
	l := landingbus.Landing{ID: uuid.New(), Label: label.MustParse("ToDelete")}
	store := newLandingStorerStub()
	store.data[l.ID] = l
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	storeErr := newLandingStorerStub()
	storeErr.deleteErr = errors.New("db error")
	busErr := landingbus.NewBusiness(storeErr, newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "removes-from-storer",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.Delete(ctx, l); err != nil {
					return err
				}
				_, exists := store.data[l.ID]
				return exists
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busErr.Delete(ctx, l) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testQueryByID() []unittest.Table {
	l := landingbus.Landing{ID: uuid.New(), Label: label.MustParse("Find me")}

	storeFound := newLandingStorerStub()
	storeFound.data[l.ID] = l
	busFound := landingbus.NewBusiness(storeFound, newLandingAttachmentQuerierStub())

	busNotFound := landingbus.NewBusiness(newLandingStorerStub(), newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-landing",
			ExpResp: l.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busFound.QueryByID(ctx, l.ID)
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
				_, err := busNotFound.QueryByID(ctx, uuid.New())
				return errors.Is(err, landingbus.ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCount() []unittest.Table {
	store := newLandingStorerStub()
	for range 4 {
		store.data[uuid.New()] = landingbus.Landing{ID: uuid.New()}
	}
	bus := landingbus.NewBusiness(store, newLandingAttachmentQuerierStub())

	storeErr := newLandingStorerStub()
	storeErr.countErr = errors.New("db error")
	busErr := landingbus.NewBusiness(storeErr, newLandingAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-correct-number",
			ExpResp: 4,
			ExcFunc: func(ctx context.Context) any {
				count, err := bus.Count(ctx, landingbus.QueryFilter{})
				if err != nil {
					return err
				}
				return count
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.Count(ctx, landingbus.QueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
