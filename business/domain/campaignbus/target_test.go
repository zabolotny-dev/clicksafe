package campaignbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// Используем стабы из campaign_test.go (тот же пакет).

func newTargetBus(cs *campaignStorerStub, ts *targetStorerStub) *campaignbus.TargetBusiness {
	return campaignbus.NewTargetBusiness(cs, ts)
}

// =============================================================================
// Target Save tests

func TestTargetSave_InDraftCampaign_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
		Label:  label.MustParse("Draft campaign"),
	}
	cs.data[cmp.ID] = cmp

	target, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmp.ID,
		EmployeeID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if target.ID == (uuid.UUID{}) {
		t.Error("Save must assign a non-zero UUID")
	}
	if target.Token == "" {
		t.Error("Save must assign a non-empty token")
	}
	if target.Status != campaignbus.Pending {
		t.Errorf("Status = %v, want %v", target.Status, campaignbus.Pending)
	}
	if _, exists := ts.data[target.ID]; !exists {
		t.Error("Target was not persisted to storer")
	}
}

func TestTargetSave_InPausedCampaign_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Paused,
	}
	cs.data[cmp.ID] = cmp

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmp.ID,
		EmployeeID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Save in Paused campaign returned error: %v", err)
	}
}

func TestTargetSave_InActiveCampaign_ReturnsErrCampaignLocked(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
	}
	cs.data[cmp.ID] = cmp

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmp.ID,
		EmployeeID: uuid.New(),
	})
	if !errors.Is(err, campaignbus.ErrCampaignLocked) {
		t.Fatalf("Save error = %v, want %v", err, campaignbus.ErrCampaignLocked)
	}
}

func TestTargetSave_CampaignNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: uuid.New(), // не существует
		EmployeeID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent campaign, got nil")
	}
}

// =============================================================================
// Target Delete tests

func TestTargetDelete_RemovesFromStorer(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: uuid.New(),
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	if err := bus.Delete(context.Background(), target); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, exists := ts.data[target.ID]; exists {
		t.Error("Target still exists in storer after Delete")
	}
}

// =============================================================================
// Target ChangeStatus tests

func TestTargetChangeStatus_PendingToSent_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Status: campaignbus.Pending,
	}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Sent); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
	if ts.data[target.ID].Status != campaignbus.Sent {
		t.Errorf("Status = %v, want %v", ts.data[target.ID].Status, campaignbus.Sent)
	}
}

func TestTargetChangeStatus_PendingToFailed_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Status: campaignbus.Pending,
	}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Failed); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_SentToOpened_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Status: campaignbus.Sent,
	}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Opened); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_PendingToCompleted_InvalidTransition(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Status: campaignbus.Pending,
	}
	ts.data[target.ID] = target

	err := bus.ChangeStatus(context.Background(), target, campaignbus.Submitted)
	if !errors.Is(err, campaignbus.ErrInvalidStatusTransition) {
		t.Fatalf("ChangeStatus error = %v, want %v", err, campaignbus.ErrInvalidStatusTransition)
	}
}

func TestTargetChangeStatus_ClickedToSubmitted_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Status: campaignbus.Clicked,
	}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Submitted); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

// =============================================================================
// Target UpdateSchedule tests

func TestTargetUpdateSchedule_WithinRange_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Draft,
		DateRange: dr,
	}
	cs.data[cmp.ID] = cmp

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	scheduledAt := now.Add(2 * time.Hour)
	updated, err := bus.UpdateSchedule(context.Background(), target, scheduledAt)
	if err != nil {
		t.Fatalf("UpdateSchedule returned error: %v", err)
	}
	if updated.ScheduledAt == nil {
		t.Fatal("ScheduledAt should be set")
	}
	if !updated.ScheduledAt.UTC().Equal(scheduledAt.UTC()) {
		t.Errorf("ScheduledAt = %v, want %v", updated.ScheduledAt, scheduledAt)
	}
}

func TestTargetUpdateSchedule_OutOfRange_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Draft,
		DateRange: dr,
	}
	cs.data[cmp.ID] = cmp

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	// Дата за пределами DateRange
	outOfRange := now.Add(72 * time.Hour)
	_, err := bus.UpdateSchedule(context.Background(), target, outOfRange)
	if !errors.Is(err, campaignbus.ErrInvalidScheduleWindow) {
		t.Fatalf("UpdateSchedule error = %v, want %v", err, campaignbus.ErrInvalidScheduleWindow)
	}
}

func TestTargetUpdateSchedule_ActiveCampaign_ReturnsErrTargetLocked(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Active, // Активная — нельзя менять расписание
		DateRange: dr,
	}
	cs.data[cmp.ID] = cmp

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	_, err := bus.UpdateSchedule(context.Background(), target, now.Add(2*time.Hour))
	if !errors.Is(err, campaignbus.ErrTargetLocked) {
		t.Fatalf("UpdateSchedule error = %v, want %v", err, campaignbus.ErrTargetLocked)
	}
}

func TestTargetUpdateSchedule_NonPendingTarget_ReturnsErrTargetLocked(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Draft,
		DateRange: dr,
	}
	cs.data[cmp.ID] = cmp

	// Таргет уже отправлен — нельзя менять расписание
	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Status:     campaignbus.Sent,
	}
	ts.data[target.ID] = target

	_, err := bus.UpdateSchedule(context.Background(), target, now.Add(2*time.Hour))
	if !errors.Is(err, campaignbus.ErrTargetLocked) {
		t.Fatalf("UpdateSchedule error = %v, want %v", err, campaignbus.ErrTargetLocked)
	}
}

// =============================================================================
// Target AutoDistribute tests

func TestTargetAutoDistribute_EvenDistribution(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	start := now.Add(time.Hour)
	end := now.Add(5 * time.Hour)
	dr, _ := date.ParseNull(start, end)

	cmpID := uuid.New()
	cmp := campaignbus.Campaign{
		ID:        cmpID,
		Status:    campaignbus.Draft,
		DateRange: dr,
	}
	cs.data[cmpID] = cmp

	// 4 таргета без расписания
	for i := 0; i < 4; i++ {
		id := uuid.New()
		ts.data[id] = campaignbus.Target{
			ID:         id,
			CampaignID: cmpID,
			Status:     campaignbus.Pending,
		}
	}

	if err := bus.AutoDistribute(context.Background(), cmpID); err != nil {
		t.Fatalf("AutoDistribute returned error: %v", err)
	}

	// Все таргеты должны получить ScheduledAt
	for _, target := range ts.data {
		if target.CampaignID != cmpID {
			continue
		}
		if target.ScheduledAt == nil {
			t.Errorf("target %v has no ScheduledAt after AutoDistribute", target.ID)
		}
	}
}

func TestTargetAutoDistribute_NoPendingTargets_NoOp(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(5*time.Hour))

	cmpID := uuid.New()
	cs.data[cmpID] = campaignbus.Campaign{
		ID:        cmpID,
		Status:    campaignbus.Draft,
		DateRange: dr,
	}
	// Таргетов нет

	if err := bus.AutoDistribute(context.Background(), cmpID); err != nil {
		t.Fatalf("AutoDistribute returned error for empty targets: %v", err)
	}
}

func TestTargetAutoDistribute_ActiveCampaign_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(5*time.Hour))

	cmpID := uuid.New()
	cs.data[cmpID] = campaignbus.Campaign{
		ID:        cmpID,
		Status:    campaignbus.Active,
		DateRange: dr,
	}

	err := bus.AutoDistribute(context.Background(), cmpID)
	if !errors.Is(err, campaignbus.ErrTargetLocked) {
		t.Fatalf("AutoDistribute error = %v, want %v", err, campaignbus.ErrTargetLocked)
	}
}

// =============================================================================
// Target PhishingURL tests

func TestPhishingURL_ReturnsCorrectURL(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	d := domain.MustParse("https://phishing.example.com")
	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Active,
		Domain: d,
	}
	cs.data[cmp.ID] = cmp

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Token:      "abc-token-123",
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	url, err := bus.PhishingURL(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("PhishingURL returned error: %v", err)
	}

	expected := "https://phishing.example.com/abc-token-123"
	if url != expected {
		t.Errorf("URL = %q, want %q", url, expected)
	}
}

func TestPhishingURL_NoDomain_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	// Кампания без домена
	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
	}
	cs.data[cmp.ID] = cmp

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Token:      "some-token",
		Status:     campaignbus.Pending,
	}
	ts.data[target.ID] = target

	_, err := bus.PhishingURL(context.Background(), target.ID)
	if !errors.Is(err, campaignbus.ErrDomainRequired) {
		t.Fatalf("PhishingURL error = %v, want %v", err, campaignbus.ErrDomainRequired)
	}
}

func TestPhishingURL_TargetNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	_, err := bus.PhishingURL(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent target, got nil")
	}
}

// =============================================================================
// Target QueryByID tests

func TestTargetQueryByID_ReturnsTarget(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{
		ID:     uuid.New(),
		Token:  "some-token",
		Status: campaignbus.Pending,
	}
	ts.data[target.ID] = target

	got, err := bus.QueryByID(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != target.ID {
		t.Errorf("ID = %v, want %v", got.ID, target.ID)
	}
}

func TestTargetQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, campaignbus.ErrTargetNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, campaignbus.ErrTargetNotFound)
	}
}

// =============================================================================
// TargetBusiness.Save — непокрытые ветки store ошибок

func TestTargetSave_EmployeeNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmpID := uuid.New()
	cs.data[cmpID] = campaignbus.Campaign{ID: cmpID, Status: campaignbus.Draft}
	ts.saveErr = campaignbus.ErrEmployeeNotFound

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmpID,
		EmployeeID: uuid.New(),
	})
	if !errors.Is(err, campaignbus.ErrEmployeeNotFound) {
		t.Fatalf("Save error = %v, want %v", err, campaignbus.ErrEmployeeNotFound)
	}
}

func TestTargetSave_AlreadyExists_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmpID := uuid.New()
	cs.data[cmpID] = campaignbus.Campaign{ID: cmpID, Status: campaignbus.Draft}
	ts.saveErr = campaignbus.ErrTargetAlreadyExists

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmpID,
		EmployeeID: uuid.New(),
	})
	if !errors.Is(err, campaignbus.ErrTargetAlreadyExists) {
		t.Fatalf("Save error = %v, want %v", err, campaignbus.ErrTargetAlreadyExists)
	}
}

func TestTargetSave_GenericStoreError_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmpID := uuid.New()
	cs.data[cmpID] = campaignbus.Campaign{ID: cmpID, Status: campaignbus.Draft}
	ts.saveErr = errors.New("db connection lost")

	_, err := bus.Save(context.Background(), campaignbus.NewTarget{
		CampaignID: cmpID,
		EmployeeID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error from store, got nil")
	}
}

// =============================================================================
// TargetBusiness.DeleteByCampaignID — непокрытые ветки

func TestTargetDeleteByCampaignID_CampaignNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	err := bus.DeleteByCampaignID(context.Background(), uuid.New())
	if !errors.Is(err, campaignbus.ErrCampaignNotFound) {
		t.Fatalf("DeleteByCampaignID error = %v, want %v", err, campaignbus.ErrCampaignNotFound)
	}
}

// =============================================================================
// isValidTargetTransition — непокрытые переходы

func TestTargetChangeStatus_PendingToClicked_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Pending}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Clicked); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_SentToReplied_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Sent}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Replied); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_SentToClicked_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Sent}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Clicked); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_OpenedToClicked_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Opened}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Clicked); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_OpenedToReplied_Success(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Opened}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Replied); err != nil {
		t.Fatalf("ChangeStatus returned error: %v", err)
	}
}

func TestTargetChangeStatus_FromTerminalState_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	// Failed — терминальный статус, никакой переход невалиден
	target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Failed}
	ts.data[target.ID] = target

	if err := bus.ChangeStatus(context.Background(), target, campaignbus.Sent); err == nil {
		t.Fatal("expected error for transition from terminal state, got nil")
	}
}
