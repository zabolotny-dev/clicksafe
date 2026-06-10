package campaignbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

// =============================================================================
// Stubs

type campaignStorerStub struct {
	data       map[uuid.UUID]campaignbus.Campaign
	saveErr    error
	updateErr  error
	deleteErr  error
	queryErr   error
	countErr   error
	expiredCmp []campaignbus.Campaign
	expiredErr error
}

func newCampaignStorerStub() *campaignStorerStub {
	return &campaignStorerStub{data: make(map[uuid.UUID]campaignbus.Campaign)}
}

func (s *campaignStorerStub) Save(_ context.Context, c campaignbus.Campaign) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[c.ID] = c
	return nil
}

func (s *campaignStorerStub) Update(_ context.Context, c campaignbus.Campaign) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.data[c.ID] = c
	return nil
}

func (s *campaignStorerStub) QueryByID(_ context.Context, id uuid.UUID) (campaignbus.Campaign, error) {
	c, ok := s.data[id]
	if !ok {
		return campaignbus.Campaign{}, campaignbus.ErrCampaignNotFound
	}
	return c, nil
}

func (s *campaignStorerStub) Query(_ context.Context, _ campaignbus.CampaignQueryFilter, _ order.By, _ page.Page) ([]campaignbus.Campaign, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	var result []campaignbus.Campaign
	for _, c := range s.data {
		result = append(result, c)
	}
	return result, nil
}

func (s *campaignStorerStub) Delete(_ context.Context, c campaignbus.Campaign) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.data, c.ID)
	return nil
}

func (s *campaignStorerStub) Count(_ context.Context, _ campaignbus.CampaignQueryFilter) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return len(s.data), nil
}

func (s *campaignStorerStub) QueryExpired(_ context.Context) ([]campaignbus.Campaign, error) {
	if s.expiredErr != nil {
		return nil, s.expiredErr
	}
	return s.expiredCmp, nil
}

type targetStorerStub struct {
	data                  map[uuid.UUID]campaignbus.Target
	saveErr               error
	updateErr             error
	deleteErr             error
	deleteByCampaignIDErr error
	queryDueErr           error
	queryCampaignErr      error
	queryErr              error
	countErr              error
}

func newTargetStorerStub() *targetStorerStub {
	return &targetStorerStub{data: make(map[uuid.UUID]campaignbus.Target)}
}

func (s *targetStorerStub) Save(_ context.Context, t campaignbus.Target) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[t.ID] = t
	return nil
}

func (s *targetStorerStub) Delete(_ context.Context, t campaignbus.Target) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.data, t.ID)
	return nil
}

func (s *targetStorerStub) Update(_ context.Context, t campaignbus.Target) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.data[t.ID] = t
	return nil
}

func (s *targetStorerStub) UpdateMany(_ context.Context, ts []campaignbus.Target) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	for _, t := range ts {
		s.data[t.ID] = t
	}
	return nil
}

func (s *targetStorerStub) QueryByID(_ context.Context, id uuid.UUID) (campaignbus.Target, error) {
	t, ok := s.data[id]
	if !ok {
		return campaignbus.Target{}, campaignbus.ErrTargetNotFound
	}
	return t, nil
}

func (s *targetStorerStub) QueryByToken(_ context.Context, token string) (campaignbus.Target, error) {
	for _, t := range s.data {
		if t.Token == token {
			return t, nil
		}
	}
	return campaignbus.Target{}, campaignbus.ErrTargetNotFound
}

func (s *targetStorerStub) DeleteByCampaignID(_ context.Context, campaignID uuid.UUID) error {
	if s.deleteByCampaignIDErr != nil {
		return s.deleteByCampaignIDErr
	}
	for id, t := range s.data {
		if t.CampaignID == campaignID {
			delete(s.data, id)
		}
	}
	return nil
}

func (s *targetStorerStub) QueryDue(_ context.Context, _ time.Time) ([]campaignbus.Target, error) {
	if s.queryDueErr != nil {
		return nil, s.queryDueErr
	}
	return nil, nil
}

func (s *targetStorerStub) Query(_ context.Context, filter campaignbus.TargetQueryFilter) ([]campaignbus.Target, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	var result []campaignbus.Target
	for _, t := range s.data {
		if filter.CampaignID != nil && t.CampaignID != *filter.CampaignID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (s *targetStorerStub) Count(_ context.Context, filter campaignbus.TargetQueryFilter) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	count := 0
	for _, t := range s.data {
		if filter.CampaignID != nil && t.CampaignID != *filter.CampaignID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		count++
	}
	return count, nil
}

type messageQuerierStub struct {
	data map[uuid.UUID]messagebus.Message
}

func newMessageQuerierStub() *messageQuerierStub {
	return &messageQuerierStub{data: make(map[uuid.UUID]messagebus.Message)}
}

func (s *messageQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (messagebus.Message, error) {
	msg, ok := s.data[id]
	if !ok {
		return messagebus.Message{}, messagebus.ErrNotFound
	}
	return msg, nil
}

type landingQuerierStub struct {
	data map[uuid.UUID]landingbus.Landing
}

func newLandingQuerierStub() *landingQuerierStub {
	return &landingQuerierStub{data: make(map[uuid.UUID]landingbus.Landing)}
}

func (s *landingQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (landingbus.Landing, error) {
	l, ok := s.data[id]
	if !ok {
		return landingbus.Landing{}, landingbus.ErrNotFound
	}
	return l, nil
}

type attachmentProviderStub struct {
	data map[uuid.UUID]attachmentbus.Attachment
}

func newAttachmentProviderStub() *attachmentProviderStub {
	return &attachmentProviderStub{data: make(map[uuid.UUID]attachmentbus.Attachment)}
}

func (s *attachmentProviderStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	a, ok := s.data[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return a, nil
}

type employeeQuerierStub struct {
	data map[uuid.UUID]employeebus.Employee
}

func newEmployeeQuerierStub() *employeeQuerierStub {
	return &employeeQuerierStub{data: make(map[uuid.UUID]employeebus.Employee)}
}

func (s *employeeQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (employeebus.Employee, error) {
	e, ok := s.data[id]
	if !ok {
		return employeebus.Employee{}, employeebus.ErrNotFound
	}
	return e, nil
}

type varsValidatorStub struct {
	missing []campaignbus.TargetMissingVars
	err     error
}

func (v *varsValidatorStub) Validate(_ context.Context, _ campaignbus.Campaign, _ []campaignbus.Target, _ []string) ([]campaignbus.TargetMissingVars, error) {
	return v.missing, v.err
}

func newCampaignBus(cs *campaignStorerStub, ts *targetStorerStub, mq *messageQuerierStub, lq *landingQuerierStub, eq *employeeQuerierStub, ap *attachmentProviderStub, vv *varsValidatorStub) *campaignbus.CampaignBusiness {
	return campaignbus.NewCampaignBusiness(cs, ts, mq, lq, eq, vv, ap)
}

func makeDateRange(start, end time.Time) date.Null {
	dr, err := date.ParseNull(start, end)
	if err != nil {
		panic(err)
	}
	return dr
}

// addScheduledTarget adds a target with a scheduled time to the target store.
func addScheduledTarget(ts *targetStorerStub, campID uuid.UUID) {
	scheduledAt := time.Now().UTC().Add(2 * time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  campID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
}

// =============================================================================

func Test_Campaign(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testCampaignSave(), "save")
	unittest.Run(t, testCampaignUpdate(), "update")
	unittest.Run(t, testCampaignDelete(), "delete")
	unittest.Run(t, testCampaignQueryByID(), "querybyid")
	unittest.Run(t, testCampaignCount(), "count")
	unittest.Run(t, testCampaignCompleteExpired(), "completeexpired")
	unittest.Run(t, testCampaignStart(), "start")
	unittest.Run(t, testCampaignStartEmailValidation(), "start-email-validation")
	unittest.Run(t, testCampaignStartMaxValidation(), "start-max-validation")
	unittest.Run(t, testCampaignPauseCancel(), "pause-cancel")
	unittest.Run(t, testCampaignSeedHelper(), "seed-helper")
}

func testCampaignSeedHelper() []unittest.Table {
	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "seed-creates-n-campaigns",
			ExpResp: 3,
			ExcFunc: func(ctx context.Context) any {
				camps, err := campaignbus.TestSeedCampaigns(ctx, 3, bus)
				if err != nil {
					return err
				}
				return len(camps)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignCount() []unittest.Table {
	cs := newCampaignStorerStub()
	for range 3 {
		id := uuid.New()
		cs.data[id] = campaignbus.Campaign{ID: id, Status: campaignbus.Draft}
	}
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	csErr := newCampaignStorerStub()
	csErr.countErr = errors.New("db error")
	busErr := newCampaignBus(csErr, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	csQueryErr := newCampaignStorerStub()
	csQueryErr.queryErr = errors.New("db error")
	busQueryErr := newCampaignBus(csQueryErr, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "returns-count",
			ExpResp: 3,
			ExcFunc: func(ctx context.Context) any {
				count, err := bus.Count(ctx, campaignbus.CampaignQueryFilter{})
				if err != nil {
					return err
				}
				return count
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "count-store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.Count(ctx, campaignbus.CampaignQueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQueryErr.Query(ctx, campaignbus.CampaignQueryFilter{}, campaignbus.DefaultOrderBy, page.MustParse("1", "10"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignCompleteExpired() []unittest.Table {
	now := time.Now().UTC()

	// Campaign with no pending targets — should complete
	csReady := newCampaignStorerStub()
	tsReady := newTargetStorerStub()
	expiredDr := makeDateRange(now.Add(-48*time.Hour), now.Add(-time.Hour))
	cReady := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Active, DateRange: expiredDr}
	csReady.expiredCmp = []campaignbus.Campaign{cReady}
	csReady.data[cReady.ID] = cReady
	busReady := newCampaignBus(csReady, tsReady, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Error fetching expired
	csExpErr := newCampaignStorerStub()
	csExpErr.expiredErr = errors.New("db error")
	busExpErr := newCampaignBus(csExpErr, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Empty expired list
	csEmpty := newCampaignStorerStub()
	busEmpty := newCampaignBus(csEmpty, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "no-expired-returns-nil",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				errs := busEmpty.CompleteExpired(ctx)
				return len(errs)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "expired-db-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				errs := busExpErr.CompleteExpired(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "expired-no-pending-targets-completes",
			ExpResp: campaignbus.Completed.String(),
			ExcFunc: func(ctx context.Context) any {
				busReady.CompleteExpired(ctx)
				return csReady.data[cReady.ID].Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

// =============================================================================

func testCampaignSave() []unittest.Table {
	csOK := newCampaignStorerStub()
	busOK := newCampaignBus(csOK, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	csErr := newCampaignStorerStub()
	csErr.saveErr = errors.New("db error")
	busErr := newCampaignBus(csErr, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "creates-draft-campaign",
			ExpResp: campaignbus.Draft.String(),
			ExcFunc: func(ctx context.Context) any {
				c, err := busOK.Save(ctx, campaignbus.NewCampaign{Label: label.MustParse("Test campaign")})
				if err != nil {
					return err
				}
				return c.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "defaults-to-email-type",
			ExpResp: campaignbus.EmailCampaign.String(),
			ExcFunc: func(ctx context.Context) any {
				c, err := busOK.Save(ctx, campaignbus.NewCampaign{Label: label.MustParse("Email campaign")})
				if err != nil {
					return err
				}
				return c.Type.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "assigns-non-zero-uuid",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				c, err := busOK.Save(ctx, campaignbus.NewCampaign{Label: label.MustParse("UUID campaign")})
				if err != nil {
					return err
				}
				return c.ID != (uuid.UUID{})
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.Save(ctx, campaignbus.NewCampaign{Label: label.MustParse("Campaign")})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignUpdate() []unittest.Table {
	original := campaignbus.Campaign{
		ID:     uuid.New(),
		Label:  label.MustParse("Old label"),
		Status: campaignbus.Draft,
		Type:   campaignbus.EmailCampaign,
	}
	activeCmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
		Type:   campaignbus.EmailCampaign,
	}

	cs := newCampaignStorerStub()
	cs.data[original.ID] = original
	cs.data[activeCmp.ID] = activeCmp
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	newLabel := label.MustParse("New label")
	maxType := campaignbus.MaxCampaign
	newDomain := domain.MustParse("https://new.com")
	now := time.Now().UTC()
	newDateRange := makeDateRange(now, now.Add(time.Hour))
	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	maxEduID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	newAttrs := map[string]string{"foo": "bar"}

	return []unittest.Table{
		{
			Name:    "changes-label",
			ExpResp: newLabel,
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Update(ctx, original, campaignbus.UpdateCampaign{Label: &newLabel})
				if err != nil {
					return err
				}
				return updated.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cannot-change-type-in-active-status",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Update(ctx, activeCmp, campaignbus.UpdateCampaign{Type: &maxType})
				return errors.Is(err, campaignbus.ErrCampaignLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cannot-change-domain-in-active-status",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Update(ctx, activeCmp, campaignbus.UpdateCampaign{Domain: &newDomain})
				return errors.Is(err, campaignbus.ErrCampaignLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "can-change-type-in-draft-status",
			ExpResp: campaignbus.MaxCampaign.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Update(ctx, original, campaignbus.UpdateCampaign{Type: &maxType})
				if err != nil {
					return err
				}
				return updated.Type.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "updates-all-fields-in-draft",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Update(ctx, original, campaignbus.UpdateCampaign{
					Type:               &maxType,
					MessageID:          &msgID,
					LandingID:          &landID,
					EducationID:        &eduID,
					MaxEducationTextID: &maxEduID,
					Label:              &newLabel,
					Domain:             &newDomain,
					DateRange:          &newDateRange,
					Attributes:         &newAttrs,
				})
				if err != nil {
					return err
				}
				return updated.Type == maxType &&
					updated.Label == newLabel &&
					updated.Attributes["foo"] == "bar"
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignDelete() []unittest.Table {
	c := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft}
	cs := newCampaignStorerStub()
	cs.data[c.ID] = c
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	csErr := newCampaignStorerStub()
	csErr.deleteErr = errors.New("db error")
	busErr := newCampaignBus(csErr, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "removes-from-storer",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.Delete(ctx, c); err != nil {
					return err
				}
				_, exists := cs.data[c.ID]
				return exists
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busErr.Delete(ctx, c) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignQueryByID() []unittest.Table {
	c := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft, Label: label.MustParse("Test")}
	cs := newCampaignStorerStub()
	cs.data[c.ID] = c
	busFound := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})
	busNotFound := newCampaignBus(newCampaignStorerStub(), newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "returns-campaign",
			ExpResp: c.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busFound.QueryByID(ctx, c.ID)
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
				return errors.Is(err, campaignbus.ErrCampaignNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignStart() []unittest.Table {
	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	expiredDr := makeDateRange(now.Add(-48*time.Hour), now.Add(-time.Hour))
	msgID := uuid.New()

	// No targets
	csNoTargets := newCampaignStorerStub()
	tsNoTargets := newTargetStorerStub()
	cNoTargets := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Draft, DateRange: dr, MessageID: &msgID}
	csNoTargets.data[cNoTargets.ID] = cNoTargets
	busNoTargets := newCampaignBus(csNoTargets, tsNoTargets, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// No date range
	csNoDR := newCampaignStorerStub()
	tsNoDR := newTargetStorerStub()
	cNoDR := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Draft, MessageID: &msgID}
	csNoDR.data[cNoDR.ID] = cNoDR
	addScheduledTarget(tsNoDR, cNoDR.ID)
	busNoDR := newCampaignBus(csNoDR, tsNoDR, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Unscheduled targets
	csUnscheduled := newCampaignStorerStub()
	tsUnscheduled := newTargetStorerStub()
	cUnscheduled := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Draft, DateRange: dr, MessageID: &msgID}
	csUnscheduled.data[cUnscheduled.ID] = cUnscheduled
	tsUnscheduled.data[uuid.New()] = campaignbus.Target{ID: uuid.New(), CampaignID: cUnscheduled.ID, Status: campaignbus.Pending}
	busUnscheduled := newCampaignBus(csUnscheduled, tsUnscheduled, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Invalid status (Completed)
	csCompleted := newCampaignStorerStub()
	cCompleted := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Completed}
	csCompleted.data[cCompleted.ID] = cCompleted
	busCompleted := newCampaignBus(csCompleted, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Expired date range
	csExpired := newCampaignStorerStub()
	tsExpired := newTargetStorerStub()
	cExpired := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Draft, DateRange: expiredDr, MessageID: &msgID}
	csExpired.data[cExpired.ID] = cExpired
	addScheduledTarget(tsExpired, cExpired.ID)
	busExpired := newCampaignBus(csExpired, tsExpired, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "no-targets-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoTargets.Start(ctx, cNoTargets)
				return errors.Is(err, campaignbus.ErrTargetsRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-date-range-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoDR.Start(ctx, cNoDR)
				return errors.Is(err, campaignbus.ErrDateRangeRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unscheduled-targets-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUnscheduled.Start(ctx, cUnscheduled)
				var unscheduled *campaignbus.ErrUnscheduledTargets
				return errors.As(err, &unscheduled)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "invalid-transition-from-completed",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busCompleted.Start(ctx, cCompleted)
				return errors.Is(err, campaignbus.ErrInvalidStatusTransition)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "expired-date-range-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busExpired.Start(ctx, cExpired)
				return errors.Is(err, campaignbus.ErrDateRangeExpired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignStartEmailValidation() []unittest.Table {
	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	d := domain.MustParse("https://phish.example.com")
	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	htmlID := uuid.New()
	landHTMLID := uuid.New()
	eduHTMLID := uuid.New()

	setupCampaign := func(cs *campaignStorerStub, ts *targetStorerStub) campaignbus.Campaign {
		c := campaignbus.Campaign{
			ID:          uuid.New(),
			Type:        campaignbus.EmailCampaign,
			Status:      campaignbus.Draft,
			DateRange:   dr,
			MessageID:   &msgID,
			LandingID:   &landID,
			EducationID: &eduID,
			Domain:      d,
		}
		cs.data[c.ID] = c
		addScheduledTarget(ts, c.ID)
		return c
	}

	// No message
	csNoMsg := newCampaignStorerStub()
	tsNoMsg := newTargetStorerStub()
	cNoMsg := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.EmailCampaign, Status: campaignbus.Draft, DateRange: dr}
	csNoMsg.data[cNoMsg.ID] = cNoMsg
	addScheduledTarget(tsNoMsg, cNoMsg.ID)
	busNoMsg := newCampaignBus(csNoMsg, tsNoMsg, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Wrong message type
	csMsgType := newCampaignStorerStub()
	tsMsgType := newTargetStorerStub()
	cMsgType := setupCampaign(csMsgType, tsMsgType)
	mqMsgType := newMessageQuerierStub()
	mqMsgType.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage}
	busMsgType := newCampaignBus(csMsgType, tsMsgType, mqMsgType, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// No HTML in message
	csMsgNoHTML := newCampaignStorerStub()
	tsMsgNoHTML := newTargetStorerStub()
	cMsgNoHTML := setupCampaign(csMsgNoHTML, tsMsgNoHTML)
	mqMsgNoHTML := newMessageQuerierStub()
	mqMsgNoHTML.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage}
	busMsgNoHTML := newCampaignBus(csMsgNoHTML, tsMsgNoHTML, mqMsgNoHTML, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Landing no HTML
	csLandNoHTML := newCampaignStorerStub()
	tsLandNoHTML := newTargetStorerStub()
	cLandNoHTML := setupCampaign(csLandNoHTML, tsLandNoHTML)
	mqLandNoHTML := newMessageQuerierStub()
	mqLandNoHTML.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage, HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true}}
	lqLandNoHTML := newLandingQuerierStub()
	lqLandNoHTML.data[landID] = landingbus.Landing{ID: landID}
	apLandNoHTML := newAttachmentProviderStub()
	apLandNoHTML.data[htmlID] = attachmentbus.Attachment{ID: htmlID}
	busLandNoHTML := newCampaignBus(csLandNoHTML, tsLandNoHTML, mqLandNoHTML, lqLandNoHTML, newEmployeeQuerierStub(), apLandNoHTML, &varsValidatorStub{})

	// Education no HTML
	csEduNoHTML := newCampaignStorerStub()
	tsEduNoHTML := newTargetStorerStub()
	cEduNoHTML := setupCampaign(csEduNoHTML, tsEduNoHTML)
	mqEduNoHTML := newMessageQuerierStub()
	mqEduNoHTML.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage, HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true}}
	lqEduNoHTML := newLandingQuerierStub()
	lqEduNoHTML.data[landID] = landingbus.Landing{ID: landID, HtmlBodyID: uuid.NullUUID{UUID: landHTMLID, Valid: true}}
	lqEduNoHTML.data[eduID] = landingbus.Landing{ID: eduID}
	apEduNoHTML := newAttachmentProviderStub()
	apEduNoHTML.data[htmlID] = attachmentbus.Attachment{ID: htmlID}
	apEduNoHTML.data[landHTMLID] = attachmentbus.Attachment{ID: landHTMLID}
	busEduNoHTML := newCampaignBus(csEduNoHTML, tsEduNoHTML, mqEduNoHTML, lqEduNoHTML, newEmployeeQuerierStub(), apEduNoHTML, &varsValidatorStub{})

	// Missing vars error
	csMissingVars := newCampaignStorerStub()
	tsMissingVars := newTargetStorerStub()
	cMissingVars := setupCampaign(csMissingVars, tsMissingVars)
	mqMissingVars := newMessageQuerierStub()
	mqMissingVars.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage, HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true}}
	lqMissingVars := newLandingQuerierStub()
	lqMissingVars.data[landID] = landingbus.Landing{ID: landID, HtmlBodyID: uuid.NullUUID{UUID: landHTMLID, Valid: true}}
	lqMissingVars.data[eduID] = landingbus.Landing{ID: eduID, HtmlBodyID: uuid.NullUUID{UUID: eduHTMLID, Valid: true}}
	apMissingVars := newAttachmentProviderStub()
	apMissingVars.data[htmlID] = attachmentbus.Attachment{ID: htmlID}
	apMissingVars.data[landHTMLID] = attachmentbus.Attachment{ID: landHTMLID}
	apMissingVars.data[eduHTMLID] = attachmentbus.Attachment{ID: eduHTMLID}
	vvMissingVars := &varsValidatorStub{
		missing: []campaignbus.TargetMissingVars{
			{TargetID: uuid.New(), EmployeeID: uuid.New(), Vars: []string{"FirstName"}},
		},
	}
	busMissingVars := newCampaignBus(csMissingVars, tsMissingVars, mqMissingVars, lqMissingVars, newEmployeeQuerierStub(), apMissingVars, vvMissingVars)

	// No domain
	csNoDomain := newCampaignStorerStub()
	tsNoDomain := newTargetStorerStub()
	cNoDomain := campaignbus.Campaign{
		ID:          uuid.New(),
		Type:        campaignbus.EmailCampaign,
		Status:      campaignbus.Draft,
		DateRange:   dr,
		MessageID:   &msgID,
		LandingID:   &landID,
		EducationID: &eduID,
	}
	csNoDomain.data[cNoDomain.ID] = cNoDomain
	addScheduledTarget(tsNoDomain, cNoDomain.ID)
	busNoDomain := newCampaignBus(csNoDomain, tsNoDomain, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_ = cMsgType
	_ = cMsgNoHTML
	_ = cLandNoHTML
	_ = cEduNoHTML
	_ = cMissingVars

	return []unittest.Table{
		{
			Name:    "no-message-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoMsg.Start(ctx, cNoMsg)
				return errors.Is(err, campaignbus.ErrMessageRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "wrong-message-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				c := csNoDomain.data[cNoDomain.ID]
				_ = c
				c2 := campaignbus.Campaign{
					ID:          uuid.New(),
					Type:        campaignbus.EmailCampaign,
					Status:      campaignbus.Draft,
					DateRange:   dr,
					MessageID:   &msgID,
					LandingID:   &landID,
					EducationID: &eduID,
					Domain:      d,
				}
				csMsgType.data[c2.ID] = c2
				addScheduledTarget(tsMsgType, c2.ID)
				_, err := busMsgType.Start(ctx, c2)
				return errors.Is(err, campaignbus.ErrMessageTypeMismatch)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-no-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busMsgNoHTML.Start(ctx, cMsgNoHTML)
				return errors.Is(err, campaignbus.ErrMessageHTMLRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "landing-no-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busLandNoHTML.Start(ctx, cLandNoHTML)
				return errors.Is(err, campaignbus.ErrLandingHTMLRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "education-no-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busEduNoHTML.Start(ctx, cEduNoHTML)
				return errors.Is(err, campaignbus.ErrEducationHTMLRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "missing-vars-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busMissingVars.Start(ctx, cMissingVars)
				var missingVars *campaignbus.ErrTargetsMissingVars
				return errors.As(err, &missingVars)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-domain-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoDomain.Start(ctx, cNoDomain)
				return errors.Is(err, campaignbus.ErrDomainRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignStartMaxValidation() []unittest.Table {
	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()
	maxEduID := uuid.New()
	maxAccID := uuid.New()
	txtID := uuid.New()
	empID := uuid.New()

	setupMax := func(cs *campaignStorerStub, ts *targetStorerStub) campaignbus.Campaign {
		c := campaignbus.Campaign{
			ID:                 uuid.New(),
			Type:               campaignbus.MaxCampaign,
			Status:             campaignbus.Draft,
			DateRange:          dr,
			MessageID:          &msgID,
			MaxEducationTextID: uuid.NullUUID{UUID: maxEduID, Valid: true},
		}
		cs.data[c.ID] = c
		addScheduledTarget(ts, c.ID)
		return c
	}

	// No message
	csNoMsg := newCampaignStorerStub()
	tsNoMsg := newTargetStorerStub()
	cNoMsg := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.MaxCampaign, Status: campaignbus.Draft, DateRange: dr}
	csNoMsg.data[cNoMsg.ID] = cNoMsg
	addScheduledTarget(tsNoMsg, cNoMsg.ID)
	busNoMsg := newCampaignBus(csNoMsg, tsNoMsg, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// No max education
	csNoEdu := newCampaignStorerStub()
	tsNoEdu := newTargetStorerStub()
	cNoEdu := campaignbus.Campaign{ID: uuid.New(), Type: campaignbus.MaxCampaign, Status: campaignbus.Draft, DateRange: dr, MessageID: &msgID}
	csNoEdu.data[cNoEdu.ID] = cNoEdu
	addScheduledTarget(tsNoEdu, cNoEdu.ID)
	busNoEdu := newCampaignBus(csNoEdu, tsNoEdu, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Wrong message type
	csMsgType := newCampaignStorerStub()
	tsMsgType := newTargetStorerStub()
	cMsgType := setupMax(csMsgType, tsMsgType)
	mqMsgType := newMessageQuerierStub()
	mqMsgType.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage}
	busMsgType := newCampaignBus(csMsgType, tsMsgType, mqMsgType, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// No max account
	csNoAcc := newCampaignStorerStub()
	tsNoAcc := newTargetStorerStub()
	cNoAcc := setupMax(csNoAcc, tsNoAcc)
	mqNoAcc := newMessageQuerierStub()
	mqNoAcc.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage}
	busNoAcc := newCampaignBus(csNoAcc, tsNoAcc, mqNoAcc, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	// Text body not txt
	csTxtType := newCampaignStorerStub()
	tsTxtType := newTargetStorerStub()
	cTxtType := setupMax(csTxtType, tsTxtType)
	mqTxtType := newMessageQuerierStub()
	mqTxtType.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage, MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true}, TextBodyID: uuid.NullUUID{UUID: txtID, Valid: true}}
	apTxtType := newAttachmentProviderStub()
	apTxtType.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Html}
	busTxtType := newCampaignBus(csTxtType, tsTxtType, mqTxtType, newLandingQuerierStub(), newEmployeeQuerierStub(), apTxtType, &varsValidatorStub{})

	// Edu text not txt
	csEduType := newCampaignStorerStub()
	tsEduType := newTargetStorerStub()
	cEduType := setupMax(csEduType, tsEduType)
	mqEduType := newMessageQuerierStub()
	mqEduType.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage, MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true}, TextBodyID: uuid.NullUUID{UUID: txtID, Valid: true}}
	apEduType := newAttachmentProviderStub()
	apEduType.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt}
	apEduType.data[maxEduID] = attachmentbus.Attachment{ID: maxEduID, Type: attachmentbus.Html}
	busEduType := newCampaignBus(csEduType, tsEduType, mqEduType, newLandingQuerierStub(), newEmployeeQuerierStub(), apEduType, &varsValidatorStub{})

	// Phone required
	csPhone := newCampaignStorerStub()
	tsPhone := newTargetStorerStub()
	cPhone := setupMax(csPhone, tsPhone)
	mqPhone := newMessageQuerierStub()
	mqPhone.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage, MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true}, TextBodyID: uuid.NullUUID{UUID: txtID, Valid: true}}
	apPhone := newAttachmentProviderStub()
	apPhone.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt}
	apPhone.data[maxEduID] = attachmentbus.Attachment{ID: maxEduID, Type: attachmentbus.Txt}
	eqPhone := newEmployeeQuerierStub()
	eqPhone.data[empID] = employeebus.Employee{ID: empID}
	// Привязываем сотрудника к таргету
	for k := range tsPhone.data {
		t2 := tsPhone.data[k]
		t2.EmployeeID = empID
		tsPhone.data[k] = t2
	}
	busPhone := newCampaignBus(csPhone, tsPhone, mqPhone, newLandingQuerierStub(), eqPhone, apPhone, &varsValidatorStub{})

	// siteEnabled - valid max with landing/edu
	now2 := time.Now().UTC()
	dr2 := makeDateRange(now2.Add(time.Hour), now2.Add(48*time.Hour))
	msgID2 := uuid.New()
	maxEduID2 := uuid.New()
	maxAccID2 := uuid.New()
	txtID5 := uuid.New()
	empID2 := uuid.New()
	landID2 := uuid.New()
	eduID2 := uuid.New()
	landHTMLID2 := uuid.New()
	eduHTMLID2 := uuid.New()
	d := domain.MustParse("https://phish.example.com")

	csSite := newCampaignStorerStub()
	tsSite := newTargetStorerStub()
	cSite := campaignbus.Campaign{
		ID:                 uuid.New(),
		Type:               campaignbus.MaxCampaign,
		Status:             campaignbus.Draft,
		DateRange:          dr2,
		MessageID:          &msgID2,
		MaxEducationTextID: uuid.NullUUID{UUID: maxEduID2, Valid: true},
		LandingID:          &landID2,
		EducationID:        &eduID2,
		Domain:             d,
	}
	csSite.data[cSite.ID] = cSite
	sch2 := now2.Add(2 * time.Hour)
	tsSite.data[uuid.New()] = campaignbus.Target{
		ID: uuid.New(), CampaignID: cSite.ID, Status: campaignbus.Pending,
		ScheduledAt: &sch2, EmployeeID: empID2,
	}
	mqSite := newMessageQuerierStub()
	mqSite.data[msgID2] = messagebus.Message{
		ID:           msgID2,
		Type:         messagebus.MaxMessage,
		MaxAccountID: uuid.NullUUID{UUID: maxAccID2, Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: txtID5, Valid: true},
	}
	lqSite := newLandingQuerierStub()
	lqSite.data[landID2] = landingbus.Landing{ID: landID2, HtmlBodyID: uuid.NullUUID{UUID: landHTMLID2, Valid: true}}
	lqSite.data[eduID2] = landingbus.Landing{ID: eduID2, HtmlBodyID: uuid.NullUUID{UUID: eduHTMLID2, Valid: true}}
	apSite := newAttachmentProviderStub()
	apSite.data[txtID5] = attachmentbus.Attachment{ID: txtID5, Type: attachmentbus.Txt}
	apSite.data[maxEduID2] = attachmentbus.Attachment{ID: maxEduID2, Type: attachmentbus.Txt}
	apSite.data[landHTMLID2] = attachmentbus.Attachment{ID: landHTMLID2, Type: attachmentbus.Html}
	apSite.data[eduHTMLID2] = attachmentbus.Attachment{ID: eduHTMLID2, Type: attachmentbus.Html}
	phone2, _ := phone.ParseNull("+79991234567")
	eqSite := newEmployeeQuerierStub()
	eqSite.data[empID2] = employeebus.Employee{ID: empID2, Phone: phone2}
	busSite := newCampaignBus(csSite, tsSite, mqSite, lqSite, eqSite, apSite, &varsValidatorStub{})

	_ = cMsgType
	_ = cNoAcc
	_ = cTxtType
	_ = cEduType
	_ = cPhone

	return []unittest.Table{
		{
			Name:    "no-message-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoMsg.Start(ctx, cNoMsg)
				return errors.Is(err, campaignbus.ErrMessageRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-max-education-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoEdu.Start(ctx, cNoEdu)
				return errors.Is(err, campaignbus.ErrMaxEducationRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "wrong-message-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busMsgType.Start(ctx, cMsgType)
				return errors.Is(err, campaignbus.ErrMessageTypeMismatch)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-max-account-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoAcc.Start(ctx, cNoAcc)
				return errors.Is(err, messagebus.ErrMaxAccountRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "text-body-not-txt-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busTxtType.Start(ctx, cTxtType)
				return errors.Is(err, messagebus.ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "edu-text-not-txt-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busEduType.Start(ctx, cEduType)
				return errors.Is(err, campaignbus.ErrMaxEducationTXTRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "phone-required-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busPhone.Start(ctx, cPhone)
				return errors.Is(err, campaignbus.ErrTargetPhoneRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "site-enabled-success-starts-active",
			ExpResp: campaignbus.Active.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := busSite.Start(ctx, cSite)
				if err != nil {
					return err.Error()
				}
				return updated.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testCampaignPauseCancel() []unittest.Table {
	activeCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Active}
	draftCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft}
	completedCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Completed}
	pausedCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Paused}

	cs := newCampaignStorerStub()
	cs.data[activeCmp.ID] = activeCmp
	cs.data[draftCmp.ID] = draftCmp
	cs.data[completedCmp.ID] = completedCmp
	cs.data[pausedCmp.ID] = pausedCmp
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	return []unittest.Table{
		{
			Name:    "pause-from-active-success",
			ExpResp: campaignbus.Paused.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Pause(ctx, activeCmp)
				if err != nil {
					return err
				}
				return updated.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pause-from-draft-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Pause(ctx, draftCmp)
				return errors.Is(err, campaignbus.ErrInvalidStatusTransition)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cancel-from-draft-success",
			ExpResp: campaignbus.Canceled.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Cancel(ctx, draftCmp)
				if err != nil {
					return err
				}
				return updated.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cancel-from-active-success",
			ExpResp: campaignbus.Canceled.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Cancel(ctx, activeCmp)
				if err != nil {
					return err
				}
				return updated.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cancel-from-paused-success",
			ExpResp: campaignbus.Canceled.String(),
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.Cancel(ctx, pausedCmp)
				if err != nil {
					return err
				}
				return updated.Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "cancel-from-completed-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Cancel(ctx, completedCmp)
				return errors.Is(err, campaignbus.ErrInvalidStatusTransition)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
