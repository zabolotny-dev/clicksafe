package campaignbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type campaignStorerStub struct {
	data       map[uuid.UUID]campaignbus.Campaign
	saveErr    error
	updateErr  error
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
	var result []campaignbus.Campaign
	for _, c := range s.data {
		result = append(result, c)
	}
	return result, nil
}

func (s *campaignStorerStub) Delete(_ context.Context, c campaignbus.Campaign) error {
	delete(s.data, c.ID)
	return nil
}

func (s *campaignStorerStub) Count(_ context.Context, _ campaignbus.CampaignQueryFilter) (int, error) {
	return len(s.data), nil
}

func (s *campaignStorerStub) QueryExpired(_ context.Context) ([]campaignbus.Campaign, error) {
	if s.expiredErr != nil {
		return nil, s.expiredErr
	}
	return s.expiredCmp, nil
}

type targetStorerStub struct {
	data      map[uuid.UUID]campaignbus.Target
	saveErr   error
	updateErr error
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
	for id, t := range s.data {
		if t.CampaignID == campaignID {
			delete(s.data, id)
		}
	}
	return nil
}

func (s *targetStorerStub) QueryDue(_ context.Context, _ time.Time) ([]campaignbus.Target, error) {
	return nil, nil
}

func (s *targetStorerStub) Query(_ context.Context, filter campaignbus.TargetQueryFilter) ([]campaignbus.Target, error) {
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

// newCampaignBus creates a CampaignBusiness with all stubs.
func newCampaignBus(
	cs *campaignStorerStub,
	ts *targetStorerStub,
	mq *messageQuerierStub,
	lq *landingQuerierStub,
	eq *employeeQuerierStub,
	ap *attachmentProviderStub,
	vv *varsValidatorStub,
) *campaignbus.CampaignBusiness {
	return campaignbus.NewCampaignBusiness(cs, ts, mq, lq, eq, vv, ap)
}

func makeDateRange(start, end time.Time) date.Null {
	dr, err := date.ParseNull(start, end)
	if err != nil {
		panic(err)
	}
	return dr
}

// =============================================================================
// Campaign Save tests

func TestCampaignSave_CreatesDraftCampaign(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp, err := bus.Save(context.Background(), campaignbus.NewCampaign{
		Label: label.MustParse("Test campaign"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if cmp.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if cmp.Status != campaignbus.Draft {
		t.Errorf("Status = %v, want %v", cmp.Status, campaignbus.Draft)
	}
	if _, exists := cs.data[cmp.ID]; !exists {
		t.Error("Campaign was not persisted to storer")
	}
}

func TestCampaignSave_DefaultsToEmailType(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp, err := bus.Save(context.Background(), campaignbus.NewCampaign{
		Label: label.MustParse("Email campaign"),
		// Type не указан — должен быть EmailCampaign
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if cmp.Type != campaignbus.EmailCampaign {
		t.Errorf("Type = %v, want %v", cmp.Type, campaignbus.EmailCampaign)
	}
}

func TestCampaignSave_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	cs.saveErr = errors.New("db error")
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Save(context.Background(), campaignbus.NewCampaign{
		Label: label.MustParse("Campaign"),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// Campaign Update tests

func TestCampaignUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	original := campaignbus.Campaign{
		ID:     uuid.New(),
		Label:  label.MustParse("Old label"),
		Status: campaignbus.Draft,
	}
	cs.data[original.ID] = original

	newLabel := label.MustParse("New label")
	updated, err := bus.Update(context.Background(), original, campaignbus.UpdateCampaign{
		Label: &newLabel,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
}

func TestCampaignUpdate_CannotChangeTypeInActiveStatus(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	activeCmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
		Type:   campaignbus.EmailCampaign,
	}
	cs.data[activeCmp.ID] = activeCmp

	maxType := campaignbus.MaxCampaign
	_, err := bus.Update(context.Background(), activeCmp, campaignbus.UpdateCampaign{
		Type: &maxType,
	})
	if !errors.Is(err, campaignbus.ErrCampaignLocked) {
		t.Fatalf("Update error = %v, want %v", err, campaignbus.ErrCampaignLocked)
	}
}

func TestCampaignUpdate_CannotChangeDomainInActiveStatus(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	activeCmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
	}
	cs.data[activeCmp.ID] = activeCmp

	d := domain.MustParse("https://example.com")
	_, err := bus.Update(context.Background(), activeCmp, campaignbus.UpdateCampaign{
		Domain: &d,
	})
	if !errors.Is(err, campaignbus.ErrCampaignLocked) {
		t.Fatalf("Update error = %v, want %v", err, campaignbus.ErrCampaignLocked)
	}
}

func TestCampaignUpdate_CanChangeTypeInDraftStatus(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	draftCmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
		Type:   campaignbus.EmailCampaign,
	}
	cs.data[draftCmp.ID] = draftCmp

	maxType := campaignbus.MaxCampaign
	updated, err := bus.Update(context.Background(), draftCmp, campaignbus.UpdateCampaign{
		Type: &maxType,
	})
	if err != nil {
		t.Fatalf("Update in Draft must allow type change, got error: %v", err)
	}
	if updated.Type != campaignbus.MaxCampaign {
		t.Errorf("Type = %v, want %v", updated.Type, campaignbus.MaxCampaign)
	}
}

func TestCampaignUpdate_UpdatesAllFields(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	draftCmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
		Type:   campaignbus.EmailCampaign,
	}
	cs.data[draftCmp.ID] = draftCmp

	maxType := campaignbus.MaxCampaign
	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	maxEduID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	newLabel := label.MustParse("New all fields")
	newDomain := domain.MustParse("https://new.com")
	now := time.Now().UTC()
	newDateRange := makeDateRange(now, now.Add(time.Hour))
	newAttrs := map[string]string{"foo": "bar"}

	updated, err := bus.Update(context.Background(), draftCmp, campaignbus.UpdateCampaign{
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
		t.Fatalf("Update returned error: %v", err)
	}

	if updated.Type != maxType {
		t.Errorf("Type = %v, want %v", updated.Type, maxType)
	}
	if *updated.MessageID != msgID {
		t.Errorf("MessageID = %v, want %v", *updated.MessageID, msgID)
	}
	if *updated.LandingID != landID {
		t.Errorf("LandingID = %v, want %v", *updated.LandingID, landID)
	}
	if *updated.EducationID != eduID {
		t.Errorf("EducationID = %v, want %v", *updated.EducationID, eduID)
	}
	if updated.MaxEducationTextID != maxEduID {
		t.Errorf("MaxEducationTextID = %v, want %v", updated.MaxEducationTextID, maxEduID)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
	if updated.Domain != newDomain {
		t.Errorf("Domain = %v, want %v", updated.Domain, newDomain)
	}
	if updated.DateRange != newDateRange {
		t.Errorf("DateRange = %v, want %v", updated.DateRange, newDateRange)
	}
	if len(updated.Attributes) != 1 || updated.Attributes["foo"] != "bar" {
		t.Errorf("Attributes = %v", updated.Attributes)
	}
}

// =============================================================================
// Campaign Start tests

func TestCampaignStart_EmailCampaign_NoMessage_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		// MessageID нет
	}
	cs.data[cmp.ID] = cmp

	// Добавляем таргет с расписанием
	scheduledAt := now.Add(2 * time.Hour)
	t1 := campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	ts.data[t1.ID] = t1

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMessageRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMessageRequired)
	}
}

func TestCampaignStart_NoTargets_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		MessageID: &msgID,
	}
	cs.data[cmp.ID] = cmp
	// Таргеты не добавляем

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrTargetsRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrTargetsRequired)
	}
}

func TestCampaignStart_UnscheduledTargets_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		MessageID: &msgID,
	}
	cs.data[cmp.ID] = cmp

	// Таргет без расписания (ScheduledAt == nil)
	t1 := campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: nil,
	}
	ts.data[t1.ID] = t1

	_, err := bus.Start(context.Background(), cmp)
	var unscheduled *campaignbus.ErrUnscheduledTargets
	if !errors.As(err, &unscheduled) {
		t.Fatalf("Start error = %v, want ErrUnscheduledTargets", err)
	}
}

func TestCampaignStart_NoDateRange_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	msgID := uuid.New()
	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		MessageID: &msgID,
		// DateRange не установлен
	}
	cs.data[cmp.ID] = cmp

	// Добавляем таргет — но Start упадёт раньше из-за DateRange
	scheduledAt := time.Now().Add(time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrDateRangeRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrDateRangeRequired)
	}
}

func TestCampaignStart_InvalidTransition_FromCompleted(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Type:   campaignbus.EmailCampaign,
		Status: campaignbus.Completed,
	}
	cs.data[cmp.ID] = cmp

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrInvalidStatusTransition) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrInvalidStatusTransition)
	}
}

// =============================================================================
// Campaign Pause tests

func TestCampaignPause_InvalidTransitionFromDraft_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
	}
	cs.data[cmp.ID] = cmp

	_, err := bus.Pause(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrInvalidStatusTransition) {
		t.Fatalf("Pause error = %v, want %v", err, campaignbus.ErrInvalidStatusTransition)
	}
}

func TestCampaignPause_FromActive_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
	}
	cs.data[cmp.ID] = cmp

	updated, err := bus.Pause(context.Background(), cmp)
	if err != nil {
		t.Fatalf("Pause returned error: %v", err)
	}
	if updated.Status != campaignbus.Paused {
		t.Errorf("Status = %v, want %v", updated.Status, campaignbus.Paused)
	}
}

// =============================================================================
// Campaign Cancel tests

func TestCampaignCancel_FromDraft_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
	}
	cs.data[cmp.ID] = cmp

	updated, err := bus.Cancel(context.Background(), cmp)
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if updated.Status != campaignbus.Canceled {
		t.Errorf("Status = %v, want %v", updated.Status, campaignbus.Canceled)
	}
}

func TestCampaignCancel_FromActive_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
	}
	cs.data[cmp.ID] = cmp

	updated, err := bus.Cancel(context.Background(), cmp)
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if updated.Status != campaignbus.Canceled {
		t.Errorf("Status = %v, want %v", updated.Status, campaignbus.Canceled)
	}
}

func TestCampaignCancel_FromCompleted_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Completed,
	}
	cs.data[cmp.ID] = cmp

	_, err := bus.Cancel(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrInvalidStatusTransition) {
		t.Fatalf("Cancel error = %v, want %v", err, campaignbus.ErrInvalidStatusTransition)
	}
}

// =============================================================================
// Campaign Delete tests

func TestCampaignDelete_RemovesFromStorer(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
	}
	cs.data[cmp.ID] = cmp

	if err := bus.Delete(context.Background(), cmp); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, exists := cs.data[cmp.ID]; exists {
		t.Error("Campaign still exists in storer after Delete")
	}
}

// =============================================================================
// Campaign QueryByID tests

func TestCampaignQueryByID_ReturnsCampaign(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
		Label:  label.MustParse("Test"),
	}
	cs.data[cmp.ID] = cmp

	got, err := bus.QueryByID(context.Background(), cmp.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != cmp.ID {
		t.Errorf("ID = %v, want %v", got.ID, cmp.ID)
	}
}

func TestCampaignQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, campaignbus.ErrCampaignNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, campaignbus.ErrCampaignNotFound)
	}
}

// =============================================================================
// MaxCampaign Start tests (validateMaxStart)

func TestCampaignStart_MaxCampaign_NoMessage_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.MaxCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		// MessageID нет
	}
	cs.data[cmp.ID] = cmp

	scheduledAt := now.Add(2 * time.Hour)
	t1 := campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	ts := newTargetStorerStub()
	ts.data[t1.ID] = t1

	bus = newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMessageRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMessageRequired)
	}
}

func TestCampaignStart_MaxCampaign_NoMaxEducation_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.MaxCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		MessageID: &msgID,
		// MaxEducationTextID нет (не Valid)
	}
	cs.data[cmp.ID] = cmp

	scheduledAt := now.Add(2 * time.Hour)
	t1 := campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	ts := newTargetStorerStub()
	ts.data[t1.ID] = t1

	bus = newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMaxEducationRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMaxEducationRequired)
	}
}

func TestCampaignStart_EmailCampaign_NoDomain_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:          uuid.New(),
		Type:        campaignbus.EmailCampaign,
		Status:      campaignbus.Draft,
		DateRange:   dr,
		MessageID:   &msgID,
		LandingID:   &landID,
		EducationID: &eduID,
		Domain:      domain.Domain{}, // пустой домен
	}
	cs.data[cmp.ID] = cmp

	scheduledAt := now.Add(2 * time.Hour)
	t1 := campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	ts.data[t1.ID] = t1

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrDomainRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrDomainRequired)
	}
}

// =============================================================================
// validateEmailStart — непокрытые ветви

func setupEmailCampaign(cs *campaignStorerStub, ts *targetStorerStub, msgID, landID, eduID uuid.UUID) campaignbus.Campaign {
	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))
	d := domain.MustParse("https://phish.example.com")
	cmp := campaignbus.Campaign{
		ID:          uuid.New(),
		Type:        campaignbus.EmailCampaign,
		Status:      campaignbus.Draft,
		DateRange:   dr,
		MessageID:   &msgID,
		LandingID:   &landID,
		EducationID: &eduID,
		Domain:      d,
	}
	cs.data[cmp.ID] = cmp
	scheduledAt := now.Add(2 * time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	return cmp
}

func TestCampaignStart_EmailCampaign_WrongMessageType_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)

	// Сообщение есть, но оно MaxMessage — не совпадает с EmailCampaign
	mq.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage}

	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMessageTypeMismatch) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMessageTypeMismatch)
	}
}

func TestCampaignStart_EmailCampaign_MessageNoHTML_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)

	// Сообщение есть, тип правильный, но HtmlBodyID не заполнен
	mq.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage}

	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMessageHTMLRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMessageHTMLRequired)
	}
}

func TestCampaignStart_EmailCampaign_LandingNoHTML_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	htmlID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)

	mq.data[msgID] = messagebus.Message{
		ID:         msgID,
		Type:       messagebus.EmailMessage,
		HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
	}
	// Лендинг есть, но без HTML
	lq.data[landID] = landingbus.Landing{ID: landID}

	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrLandingHTMLRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrLandingHTMLRequired)
	}
}

func TestCampaignStart_EmailCampaign_EducationNoHTML_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	htmlID := uuid.New()
	landHTMLID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)

	mq.data[msgID] = messagebus.Message{
		ID:         msgID,
		Type:       messagebus.EmailMessage,
		HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
	}
	lq.data[landID] = landingbus.Landing{
		ID:         landID,
		HtmlBodyID: uuid.NullUUID{UUID: landHTMLID, Valid: true},
	}
	// Образование без HTML
	lq.data[eduID] = landingbus.Landing{ID: eduID}

	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrEducationHTMLRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrEducationHTMLRequired)
	}
}

func TestCampaignStart_EmailCampaign_VarsValidatorError_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()
	ap := newAttachmentProviderStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	htmlID := uuid.New()
	landHTMLID := uuid.New()
	eduHTMLID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)

	mq.data[msgID] = messagebus.Message{
		ID:         msgID,
		Type:       messagebus.EmailMessage,
		HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
	}
	lq.data[landID] = landingbus.Landing{ID: landID, HtmlBodyID: uuid.NullUUID{UUID: landHTMLID, Valid: true}}
	lq.data[eduID] = landingbus.Landing{ID: eduID, HtmlBodyID: uuid.NullUUID{UUID: eduHTMLID, Valid: true}}
	ap.data[htmlID] = attachmentbus.Attachment{ID: htmlID}
	ap.data[landHTMLID] = attachmentbus.Attachment{ID: landHTMLID}
	ap.data[eduHTMLID] = attachmentbus.Attachment{ID: eduHTMLID}

	vv := &varsValidatorStub{err: errors.New("validator error")}
	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), ap, vv)

	_, err := bus.Start(context.Background(), cmp)
	if err == nil {
		t.Fatal("expected error from vars validator, got nil")
	}
}

func TestCampaignStart_EmailCampaign_MissingVars_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()
	ap := newAttachmentProviderStub()

	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()
	htmlID := uuid.New()
	landHTMLID := uuid.New()
	eduHTMLID := uuid.New()
	cmp := setupEmailCampaign(cs, ts, msgID, landID, eduID)
	targetID := uuid.New()

	mq.data[msgID] = messagebus.Message{
		ID:         msgID,
		Type:       messagebus.EmailMessage,
		HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
	}
	lq.data[landID] = landingbus.Landing{ID: landID, HtmlBodyID: uuid.NullUUID{UUID: landHTMLID, Valid: true}}
	lq.data[eduID] = landingbus.Landing{ID: eduID, HtmlBodyID: uuid.NullUUID{UUID: eduHTMLID, Valid: true}}
	ap.data[htmlID] = attachmentbus.Attachment{ID: htmlID}
	ap.data[landHTMLID] = attachmentbus.Attachment{ID: landHTMLID}
	ap.data[eduHTMLID] = attachmentbus.Attachment{ID: eduHTMLID}

	vv := &varsValidatorStub{
		missing: []campaignbus.TargetMissingVars{
			{TargetID: targetID, EmployeeID: uuid.New(), Vars: []string{"FirstName"}},
		},
	}
	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), ap, vv)

	_, err := bus.Start(context.Background(), cmp)
	var missingVars *campaignbus.ErrTargetsMissingVars
	if !errors.As(err, &missingVars) {
		t.Fatalf("Start error = %v, want ErrTargetsMissingVars", err)
	}
}

// =============================================================================
// validateMaxStart — непокрытые ветви

func setupMaxCampaign(cs *campaignStorerStub, ts *targetStorerStub, msgID uuid.UUID, maxEduID uuid.UUID) campaignbus.Campaign {
	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))
	cmp := campaignbus.Campaign{
		ID:                 uuid.New(),
		Type:               campaignbus.MaxCampaign,
		Status:             campaignbus.Draft,
		DateRange:          dr,
		MessageID:          &msgID,
		MaxEducationTextID: uuid.NullUUID{UUID: maxEduID, Valid: true},
	}
	cs.data[cmp.ID] = cmp
	scheduledAt := now.Add(2 * time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}
	return cmp
}

func TestCampaignStart_MaxCampaign_WrongMessageType_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	// Сообщение EmailMessage — не совпадает с MaxCampaign
	mq.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.EmailMessage}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMessageTypeMismatch) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMessageTypeMismatch)
	}
}

func TestCampaignStart_MaxCampaign_NoMaxAccount_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	mq.data[msgID] = messagebus.Message{ID: msgID, Type: messagebus.MaxMessage}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, messagebus.ErrMaxAccountRequired) {
		t.Fatalf("Start error = %v, want %v", err, messagebus.ErrMaxAccountRequired)
	}
}

func TestCampaignStart_MaxCampaign_NoTextBody_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	maxAccID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	mq.data[msgID] = messagebus.Message{
		ID:           msgID,
		Type:         messagebus.MaxMessage,
		MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true},
	}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, messagebus.ErrTextBodyRequired) {
		t.Fatalf("Start error = %v, want %v", err, messagebus.ErrTextBodyRequired)
	}
}

func TestCampaignStart_MaxCampaign_TextBodyNotTxt_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	ap := newAttachmentProviderStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	maxAccID := uuid.New()
	txtID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	mq.data[msgID] = messagebus.Message{
		ID:           msgID,
		Type:         messagebus.MaxMessage,
		MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: txtID, Valid: true},
	}
	// TextBody есть, но это HTML, не TXT
	ap.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Html}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), newEmployeeQuerierStub(), ap, &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, messagebus.ErrInvalidAttachment) {
		t.Fatalf("Start error = %v, want %v", err, messagebus.ErrInvalidAttachment)
	}
}

func TestCampaignStart_MaxCampaign_EduTextNotTxt_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	ap := newAttachmentProviderStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	maxAccID := uuid.New()
	txtID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	mq.data[msgID] = messagebus.Message{
		ID:           msgID,
		Type:         messagebus.MaxMessage,
		MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: txtID, Valid: true},
	}
	ap.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt}
	// MaxEducationText есть, но это не TXT
	ap.data[maxEduID] = attachmentbus.Attachment{ID: maxEduID, Type: attachmentbus.Html}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), newEmployeeQuerierStub(), ap, &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrMaxEducationTXTRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrMaxEducationTXTRequired)
	}
}

func TestCampaignStart_MaxCampaign_PhoneRequired_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	ap := newAttachmentProviderStub()
	eq := newEmployeeQuerierStub()

	msgID := uuid.New()
	maxEduID := uuid.New()
	maxAccID := uuid.New()
	txtID := uuid.New()
	empID := uuid.New()
	cmp := setupMaxCampaign(cs, ts, msgID, maxEduID)

	// Привязываем конкретного сотрудника к таргету
	now := time.Now().UTC()
	scheduledAt := now.Add(2 * time.Hour)
	for k := range ts.data {
		t2 := ts.data[k]
		t2.EmployeeID = empID
		ts.data[k] = t2
		_ = scheduledAt
	}

	mq.data[msgID] = messagebus.Message{
		ID:           msgID,
		Type:         messagebus.MaxMessage,
		MaxAccountID: uuid.NullUUID{UUID: maxAccID, Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: txtID, Valid: true},
	}
	ap.data[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt}
	ap.data[maxEduID] = attachmentbus.Attachment{ID: maxEduID, Type: attachmentbus.Txt}

	// Сотрудник без телефона
	eq.data[empID] = employeebus.Employee{ID: empID}

	bus := newCampaignBus(cs, ts, mq, newLandingQuerierStub(), eq, ap, &varsValidatorStub{})

	_, err := bus.Start(context.Background(), cmp)
	if !errors.Is(err, campaignbus.ErrTargetPhoneRequired) {
		t.Fatalf("Start error = %v, want %v", err, campaignbus.ErrTargetPhoneRequired)
	}
}
