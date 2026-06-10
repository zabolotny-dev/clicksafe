package campaignbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func newTargetBus(cs *campaignStorerStub, ts *targetStorerStub) *campaignbus.TargetBusiness {
	return campaignbus.NewTargetBusiness(cs, ts)
}

func Test_Target(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testTargetSave(), "save")
	unittest.Run(t, testTargetDelete(), "delete")
	unittest.Run(t, testTargetChangeStatus(), "changestatus")
	unittest.Run(t, testTargetUpdateSchedule(), "updateschedule")
	unittest.Run(t, testTargetAutoDistribute(), "autodistribute")
	unittest.Run(t, testTargetPhishingURL(), "phishingurl")
	unittest.Run(t, testTargetQueryByID(), "querybyid")
	unittest.Run(t, testTargetDeleteByCampaignID(), "deletebycampaignid")
	unittest.Run(t, testTargetQueryDue(), "querydue")
	unittest.Run(t, testTargetQueryCampaignByID(), "querycampaignbyid")
}

// =============================================================================

func testTargetSave() []unittest.Table {
	csDraft := newCampaignStorerStub()
	csDraft.data[uuid.Nil] = campaignbus.Campaign{} // placeholder
	draftCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft, Label: label.MustParse("Draft campaign")}
	pausedCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Paused, Label: label.MustParse("Paused")}
	activeCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Active, Label: label.MustParse("Active")}

	cs := newCampaignStorerStub()
	cs.data[draftCmp.ID] = draftCmp
	cs.data[pausedCmp.ID] = pausedCmp
	cs.data[activeCmp.ID] = activeCmp

	ts := newTargetStorerStub()
	bus := newTargetBus(cs, ts)

	tsEmpErr := newTargetStorerStub()
	tsEmpErr.saveErr = campaignbus.ErrEmployeeNotFound
	busEmpErr := newTargetBus(cs, tsEmpErr)

	tsExistsErr := newTargetStorerStub()
	tsExistsErr.saveErr = campaignbus.ErrTargetAlreadyExists
	busExistsErr := newTargetBus(cs, tsExistsErr)

	tsGenErr := newTargetStorerStub()
	tsGenErr.saveErr = errors.New("db connection lost")
	busGenErr := newTargetBus(cs, tsGenErr)

	csEmpty := newCampaignStorerStub()
	busNotFound := newTargetBus(csEmpty, newTargetStorerStub())

	return []unittest.Table{
		{
			Name:    "in-draft-campaign-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				t2, err := bus.Save(ctx, campaignbus.NewTarget{CampaignID: draftCmp.ID, EmployeeID: uuid.New()})
				if err != nil {
					return err
				}
				return t2.ID != (uuid.UUID{}) && t2.Token != "" && t2.Status == campaignbus.Pending
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "in-paused-campaign-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Save(ctx, campaignbus.NewTarget{CampaignID: pausedCmp.ID, EmployeeID: uuid.New()})
				return err == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "in-active-campaign-returns-locked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.Save(ctx, campaignbus.NewTarget{CampaignID: activeCmp.ID, EmployeeID: uuid.New()})
				return errors.Is(err, campaignbus.ErrCampaignLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNotFound.Save(ctx, campaignbus.NewTarget{CampaignID: uuid.New(), EmployeeID: uuid.New()})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busEmpErr.Save(ctx, campaignbus.NewTarget{CampaignID: draftCmp.ID, EmployeeID: uuid.New()})
				return errors.Is(err, campaignbus.ErrEmployeeNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-already-exists-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busExistsErr.Save(ctx, campaignbus.NewTarget{CampaignID: draftCmp.ID, EmployeeID: uuid.New()})
				return errors.Is(err, campaignbus.ErrTargetAlreadyExists)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "generic-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busGenErr.Save(ctx, campaignbus.NewTarget{CampaignID: draftCmp.ID, EmployeeID: uuid.New()})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetDelete() []unittest.Table {
	target := campaignbus.Target{ID: uuid.New(), CampaignID: uuid.New(), Status: campaignbus.Pending}
	ts := newTargetStorerStub()
	ts.data[target.ID] = target
	bus := newTargetBus(newCampaignStorerStub(), ts)

	tsErr := newTargetStorerStub()
	tsErr.deleteErr = errors.New("db error")
	busErr := newTargetBus(newCampaignStorerStub(), tsErr)

	return []unittest.Table{
		{
			Name:    "removes-from-storer",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.Delete(ctx, target); err != nil {
					return err
				}
				_, exists := ts.data[target.ID]
				return exists
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busErr.Delete(ctx, target) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetChangeStatus() []unittest.Table {
	makeTarget := func(status campaignbus.TargetStatus) campaignbus.Target {
		return campaignbus.Target{ID: uuid.New(), Status: status}
	}

	buildBus := func(t campaignbus.Target) (*targetStorerStub, *campaignbus.TargetBusiness) {
		ts := newTargetStorerStub()
		ts.data[t.ID] = t
		return ts, newTargetBus(newCampaignStorerStub(), ts)
	}

	pendingTarget := makeTarget(campaignbus.Pending)
	pendingSent := makeTarget(campaignbus.Pending)
	pendingFailed := makeTarget(campaignbus.Pending)
	pendingClicked := makeTarget(campaignbus.Pending)
	sentTarget := makeTarget(campaignbus.Sent)
	sentOpened := makeTarget(campaignbus.Sent)
	sentClicked := makeTarget(campaignbus.Sent)
	sentReplied := makeTarget(campaignbus.Sent)
	openedClicked := makeTarget(campaignbus.Opened)
	openedReplied := makeTarget(campaignbus.Opened)
	clickedSubmitted := makeTarget(campaignbus.Clicked)
	clickedReplied := makeTarget(campaignbus.Clicked)
	failedTarget := makeTarget(campaignbus.Failed)
	pendingInvalid := makeTarget(campaignbus.Pending)

	tsPendingSent, busPendingSent := buildBus(pendingSent)
	_, busPendingFailed := buildBus(pendingFailed)
	_, busPendingClicked := buildBus(pendingClicked)
	_, busSentOpened := buildBus(sentOpened)
	_, busSentClicked := buildBus(sentClicked)
	_, busSentReplied := buildBus(sentReplied)
	_, busOpenedClicked := buildBus(openedClicked)
	_, busOpenedReplied := buildBus(openedReplied)
	_, busClickedSubmitted := buildBus(clickedSubmitted)
	_, busClickedReplied := buildBus(clickedReplied)
	_, busFailedTarget := buildBus(failedTarget)
	_, busPendingInvalid := buildBus(pendingInvalid)

	_ = tsPendingSent
	_ = pendingTarget
	_ = sentTarget

	return []unittest.Table{
		{
			Name:    "pending-to-sent",
			ExpResp: campaignbus.Sent.String(),
			ExcFunc: func(ctx context.Context) any {
				if err := busPendingSent.ChangeStatus(ctx, pendingSent, campaignbus.Sent); err != nil {
					return err
				}
				return tsPendingSent.data[pendingSent.ID].Status.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pending-to-failed",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busPendingFailed.ChangeStatus(ctx, pendingFailed, campaignbus.Failed) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pending-to-clicked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busPendingClicked.ChangeStatus(ctx, pendingClicked, campaignbus.Clicked) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "sent-to-opened",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busSentOpened.ChangeStatus(ctx, sentOpened, campaignbus.Opened) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "sent-to-clicked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busSentClicked.ChangeStatus(ctx, sentClicked, campaignbus.Clicked) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "sent-to-replied",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busSentReplied.ChangeStatus(ctx, sentReplied, campaignbus.Replied) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "opened-to-clicked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busOpenedClicked.ChangeStatus(ctx, openedClicked, campaignbus.Clicked) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "opened-to-replied",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busOpenedReplied.ChangeStatus(ctx, openedReplied, campaignbus.Replied) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "clicked-to-submitted",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busClickedSubmitted.ChangeStatus(ctx, clickedSubmitted, campaignbus.Submitted) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "clicked-to-replied",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busClickedReplied.ChangeStatus(ctx, clickedReplied, campaignbus.Replied) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "invalid-transition-from-failed",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busFailedTarget.ChangeStatus(ctx, failedTarget, campaignbus.Sent) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pending-to-submitted-invalid",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busPendingInvalid.ChangeStatus(ctx, pendingInvalid, campaignbus.Submitted)
				return errors.Is(err, campaignbus.ErrInvalidStatusTransition)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-update-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tsUpErr := newTargetStorerStub()
				target := campaignbus.Target{ID: uuid.New(), Status: campaignbus.Pending}
				tsUpErr.data[target.ID] = target
				tsUpErr.updateErr = errors.New("db error")
				busUpErr := newTargetBus(newCampaignStorerStub(), tsUpErr)
				return busUpErr.ChangeStatus(ctx, target, campaignbus.Sent) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetUpdateSchedule() []unittest.Table {
	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(48*time.Hour))

	draftCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft, DateRange: dr}
	activeCmp := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Active, DateRange: dr}

	pendingTarget := campaignbus.Target{ID: uuid.New(), CampaignID: draftCmp.ID, Status: campaignbus.Pending}
	sentTarget := campaignbus.Target{ID: uuid.New(), CampaignID: draftCmp.ID, Status: campaignbus.Sent}
	pendingActiveTarget := campaignbus.Target{ID: uuid.New(), CampaignID: activeCmp.ID, Status: campaignbus.Pending}

	cs := newCampaignStorerStub()
	cs.data[draftCmp.ID] = draftCmp
	cs.data[activeCmp.ID] = activeCmp

	ts := newTargetStorerStub()
	ts.data[pendingTarget.ID] = pendingTarget
	ts.data[sentTarget.ID] = sentTarget
	ts.data[pendingActiveTarget.ID] = pendingActiveTarget

	bus := newTargetBus(cs, ts)

	withinRange := now.Add(2 * time.Hour)
	outOfRange := now.Add(72 * time.Hour)

	return []unittest.Table{
		{
			Name:    "within-range-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				updated, err := bus.UpdateSchedule(ctx, pendingTarget, withinRange)
				if err != nil {
					return err
				}
				return updated.ScheduledAt != nil && updated.ScheduledAt.UTC().Equal(withinRange.UTC())
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "out-of-range-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.UpdateSchedule(ctx, pendingTarget, outOfRange)
				return errors.Is(err, campaignbus.ErrInvalidScheduleWindow)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "active-campaign-returns-target-locked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.UpdateSchedule(ctx, pendingActiveTarget, withinRange)
				return errors.Is(err, campaignbus.ErrTargetLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "non-pending-target-returns-target-locked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.UpdateSchedule(ctx, sentTarget, withinRange)
				return errors.Is(err, campaignbus.ErrTargetLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-update-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tsUpErr := newTargetStorerStub()
				tsUpErr.data[pendingTarget.ID] = pendingTarget
				tsUpErr.updateErr = errors.New("db error")
				busUpErr := newTargetBus(cs, tsUpErr)
				_, err := busUpErr.UpdateSchedule(ctx, pendingTarget, withinRange)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetAutoDistribute() []unittest.Table {
	now := time.Now().UTC()
	dr, _ := date.ParseNull(now.Add(time.Hour), now.Add(5*time.Hour))

	cmpID := uuid.New()
	activeCmpID := uuid.New()

	cs := newCampaignStorerStub()
	cs.data[cmpID] = campaignbus.Campaign{ID: cmpID, Status: campaignbus.Draft, DateRange: dr}
	cs.data[activeCmpID] = campaignbus.Campaign{ID: activeCmpID, Status: campaignbus.Active, DateRange: dr}

	ts := newTargetStorerStub()
	for range 4 {
		id := uuid.New()
		ts.data[id] = campaignbus.Target{ID: id, CampaignID: cmpID, Status: campaignbus.Pending}
	}

	bus := newTargetBus(cs, ts)

	return []unittest.Table{
		{
			Name:    "even-distribution",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.AutoDistribute(ctx, cmpID); err != nil {
					return err
				}
				for _, t := range ts.data {
					if t.CampaignID != cmpID {
						continue
					}
					if t.ScheduledAt == nil {
						return false
					}
				}
				return true
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "active-campaign-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := bus.AutoDistribute(ctx, activeCmpID)
				return errors.Is(err, campaignbus.ErrTargetLocked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-pending-targets-noop",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				emptyID := uuid.New()
				cs.data[emptyID] = campaignbus.Campaign{ID: emptyID, Status: campaignbus.Draft, DateRange: dr}
				return bus.AutoDistribute(ctx, emptyID) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetPhishingURL() []unittest.Table {
	d := domain.MustParse("https://phishing.example.com")
	cmpWithDomain := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Active, Domain: d}
	cmpNoDomain := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft}

	target := campaignbus.Target{ID: uuid.New(), CampaignID: cmpWithDomain.ID, Token: "abc-token-123", Status: campaignbus.Pending}
	targetNoDomain := campaignbus.Target{ID: uuid.New(), CampaignID: cmpNoDomain.ID, Token: "some-token", Status: campaignbus.Pending}

	cs := newCampaignStorerStub()
	cs.data[cmpWithDomain.ID] = cmpWithDomain
	cs.data[cmpNoDomain.ID] = cmpNoDomain

	ts := newTargetStorerStub()
	ts.data[target.ID] = target
	ts.data[targetNoDomain.ID] = targetNoDomain

	bus := newTargetBus(cs, ts)
	busEmpty := newTargetBus(newCampaignStorerStub(), newTargetStorerStub())

	return []unittest.Table{
		{
			Name:    "returns-correct-url",
			ExpResp: "https://phishing.example.com/abc-token-123",
			ExcFunc: func(ctx context.Context) any {
				url, err := bus.PhishingURL(ctx, target.ID)
				if err != nil {
					return err
				}
				return url
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-domain-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := bus.PhishingURL(ctx, targetNoDomain.ID)
				return errors.Is(err, campaignbus.ErrDomainRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busEmpty.PhishingURL(ctx, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetQueryByID() []unittest.Table {
	target := campaignbus.Target{ID: uuid.New(), Token: "some-token", Status: campaignbus.Pending}

	ts := newTargetStorerStub()
	ts.data[target.ID] = target
	busFound := newTargetBus(newCampaignStorerStub(), ts)
	busNotFound := newTargetBus(newCampaignStorerStub(), newTargetStorerStub())

	return []unittest.Table{
		{
			Name:    "returns-target",
			ExpResp: target.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busFound.QueryByID(ctx, target.ID)
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
				return errors.Is(err, campaignbus.ErrTargetNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetDeleteByCampaignID() []unittest.Table {
	cmpID := uuid.New()

	cs := newCampaignStorerStub()
	cs.data[cmpID] = campaignbus.Campaign{ID: cmpID, Status: campaignbus.Draft}
	bus := newTargetBus(cs, newTargetStorerStub())

	busEmpty := newTargetBus(newCampaignStorerStub(), newTargetStorerStub())

	tsDeleteErr := newTargetStorerStub()
	tsDeleteErr.deleteByCampaignIDErr = errors.New("db error")
	busDeleteErr := newTargetBus(cs, tsDeleteErr)

	return []unittest.Table{
		{
			Name:    "campaign-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busEmpty.DeleteByCampaignID(ctx, uuid.New())
				return errors.Is(err, campaignbus.ErrCampaignNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "existing-campaign-success",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return bus.DeleteByCampaignID(ctx, cmpID) == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-delete-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDeleteErr.DeleteByCampaignID(ctx, cmpID) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetQueryDue() []unittest.Table {
	ts := newTargetStorerStub()
	busOK := newTargetBus(newCampaignStorerStub(), ts)

	tsErr := newTargetStorerStub()
	tsErr.queryDueErr = errors.New("db error")
	busErr := newTargetBus(newCampaignStorerStub(), tsErr)

	return []unittest.Table{
		{
			Name:    "empty-returns-nil",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				targets, err := busOK.QueryDue(ctx)
				if err != nil {
					return err
				}
				return len(targets)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "store-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.QueryDue(ctx)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTargetQueryCampaignByID() []unittest.Table {
	c := campaignbus.Campaign{ID: uuid.New(), Status: campaignbus.Draft}
	cs := newCampaignStorerStub()
	cs.data[c.ID] = c
	bus := newTargetBus(cs, newTargetStorerStub())
	busEmpty := newTargetBus(newCampaignStorerStub(), newTargetStorerStub())

	return []unittest.Table{
		{
			Name:    "found-returns-campaign",
			ExpResp: c.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := bus.QueryCampaignByID(ctx, c.ID)
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
				_, err := busEmpty.QueryCampaignByID(ctx, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
