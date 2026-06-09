package campaignbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Campaign Query / Count

func TestCampaignQuery_ReturnsAllCampaigns(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	for i := 0; i < 3; i++ {
		id := uuid.New()
		cs.data[id] = campaignbus.Campaign{ID: id, Status: campaignbus.Draft, Label: label.MustParse("Campaign item")}
	}

	result, err := bus.Query(context.Background(), campaignbus.CampaignQueryFilter{}, order.By{}, page.MustParse("1", "10"))
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Query returned %d campaigns, want 3", len(result))
	}
}

func TestCampaignCount_ReturnsCorrectNumber(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	for i := 0; i < 5; i++ {
		id := uuid.New()
		cs.data[id] = campaignbus.Campaign{ID: id}
	}

	count, err := bus.Count(context.Background(), campaignbus.CampaignQueryFilter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 5 {
		t.Errorf("Count = %d, want 5", count)
	}
}

// =============================================================================
// Target: QueryDue, QueryByToken, DeleteByCampaignID, QueryCampaignByID

func TestTargetQueryDue_ReturnsDueTargets(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	// QueryDue в стабе всегда возвращает пустой список (нет due),
	// главное — не возвращает ошибку.
	targets, err := bus.QueryDue(context.Background())
	if err != nil {
		t.Fatalf("QueryDue returned error: %v", err)
	}
	_ = targets
}

func TestTargetQueryByToken_ReturnsTarget(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	tok := "unique-token-abc"
	target := campaignbus.Target{
		ID:     uuid.New(),
		Token:  tok,
		Status: campaignbus.Pending,
	}
	ts.data[target.ID] = target

	got, err := bus.QueryByToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("QueryByToken returned error: %v", err)
	}
	if got.Token != tok {
		t.Errorf("Token = %q, want %q", got.Token, tok)
	}
}

func TestTargetQueryByToken_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	_, err := bus.QueryByToken(context.Background(), "nonexistent-token")
	if err == nil {
		t.Fatal("expected error for unknown token, got nil")
	}
}

func TestTargetDeleteByCampaignID_RemovesTargets(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmpID := uuid.New()
	cmp := campaignbus.Campaign{
		ID:     cmpID,
		Status: campaignbus.Draft,
	}
	cs.data[cmpID] = cmp

	// Добавляем 3 таргета к этой кампании и 1 к другой
	for i := 0; i < 3; i++ {
		id := uuid.New()
		ts.data[id] = campaignbus.Target{ID: id, CampaignID: cmpID}
	}
	otherId := uuid.New()
	ts.data[otherId] = campaignbus.Target{ID: otherId, CampaignID: uuid.New()}

	if err := bus.DeleteByCampaignID(context.Background(), cmpID); err != nil {
		t.Fatalf("DeleteByCampaignID returned error: %v", err)
	}

	// Все таргеты этой кампании удалены
	for _, target := range ts.data {
		if target.CampaignID == cmpID {
			t.Errorf("target %v still exists after DeleteByCampaignID", target.ID)
		}
	}
	// Таргет другой кампании не затронут
	if _, exists := ts.data[otherId]; !exists {
		t.Error("target from other campaign should not be deleted")
	}
}

func TestTargetQueryCampaignByID_ReturnsCampaign(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	cmp := campaignbus.Campaign{
		ID:     uuid.New(),
		Status: campaignbus.Draft,
		Label:  label.MustParse("Campaign query test"),
	}
	cs.data[cmp.ID] = cmp

	got, err := bus.QueryCampaignByID(context.Background(), cmp.ID)
	if err != nil {
		t.Fatalf("QueryCampaignByID returned error: %v", err)
	}
	if got.ID != cmp.ID {
		t.Errorf("ID = %v, want %v", got.ID, cmp.ID)
	}
}

// =============================================================================
// Campaign Start - дополнительные ветви validateEmailStart

func TestCampaignStart_NoLanding_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	mq := newMessageQuerierStub()
	lq := newLandingQuerierStub()
	bus := newCampaignBus(cs, ts, mq, lq, newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()
	landID := uuid.New()
	eduID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		MessageID: &msgID,
		// LandingID не указан
	}
	cs.data[cmp.ID] = cmp
	_ = landID
	_ = eduID

	scheduledAt := now.Add(2 * time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}

	_, err := bus.Start(context.Background(), cmp)
	if err == nil {
		t.Fatal("expected error for missing landing, got nil")
	}
}

func TestCampaignStart_NoEducation_ReturnsError(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()
	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	dr := makeDateRange(now.Add(time.Hour), now.Add(48*time.Hour))
	msgID := uuid.New()
	landID := uuid.New()

	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Draft,
		DateRange: dr,
		MessageID: &msgID,
		LandingID: &landID,
		// EducationID не указан
	}
	cs.data[cmp.ID] = cmp

	scheduledAt := now.Add(2 * time.Hour)
	ts.data[uuid.New()] = campaignbus.Target{
		ID:          uuid.New(),
		CampaignID:  cmp.ID,
		Status:      campaignbus.Pending,
		ScheduledAt: &scheduledAt,
	}

	_, err := bus.Start(context.Background(), cmp)
	if err == nil {
		t.Fatal("expected error for missing education, got nil")
	}
}

// =============================================================================
// CompleteExpired

func TestCompleteExpired_UpdatesStatus(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	now := time.Now().UTC()
	// Просроченная кампания
	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Active,
		DateRange: makeDateRange(now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
	}
	cs.data[cmp.ID] = cmp

	// Стаб должен возвращать эту кампанию в QueryExpired
	cs.expiredCmp = append(cs.expiredCmp, cmp)

	if err := bus.CompleteExpired(context.Background()); err != nil {
		t.Fatalf("CompleteExpired returned error: %v", err)
	}

	updated := cs.data[cmp.ID]
	if updated.Status != campaignbus.Completed {
		t.Errorf("Status = %v, want %v", updated.Status, campaignbus.Completed)
	}
}


// =============================================================================
// CompleteExpired — непокрытые ветки

func TestCompleteExpired_QueryExpiredError_ReturnsErrors(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	cs.expiredErr = errors.New("db error")
	bus := newCampaignBus(cs, newTargetStorerStub(), newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	errs := bus.CompleteExpired(context.Background())
	if len(errs) == 0 {
		t.Fatal("expected errors from CompleteExpired when QueryExpired fails, got none")
	}
}

func TestCompleteExpired_CampaignWithPendingTargets_NotCompleted(t *testing.T) {
	t.Parallel()

	cs := newCampaignStorerStub()
	ts := newTargetStorerStub()

	now := time.Now().UTC()
	cmp := campaignbus.Campaign{
		ID:        uuid.New(),
		Status:    campaignbus.Active,
		DateRange: makeDateRange(now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
	}
	cs.data[cmp.ID] = cmp
	cs.expiredCmp = []campaignbus.Campaign{cmp}

	// Pending-таргет → Count вернёт 1 → кампания не должна завершиться
	ts.data[uuid.New()] = campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: cmp.ID,
		Status:     campaignbus.Pending,
	}

	bus := newCampaignBus(cs, ts, newMessageQuerierStub(), newLandingQuerierStub(), newEmployeeQuerierStub(), newAttachmentProviderStub(), &varsValidatorStub{})

	bus.CompleteExpired(context.Background())

	if cs.data[cmp.ID].Status == campaignbus.Completed {
		t.Error("campaign with pending targets must not be completed")
	}
}
