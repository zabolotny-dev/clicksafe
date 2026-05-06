package targetbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

var (
	ErrNotFound                = errors.New("target not found")
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrCampaignNotFound        = errors.New("campaign not found")
	ErrCampaignNotDraft        = errors.New("campaign is not in draft status")
	ErrInvalidStatusValue      = errors.New("invalid status")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrTargetNotPending        = errors.New("target is not in pending status")
	ErrInvalidScheduleWindow   = errors.New("invalid campaign schedule window")
	ErrUniqueToken             = errors.New("unique token constraint violation")
	ErrTargetAlreadyExists     = errors.New("target already exists")
)

type CampaignQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error)
}

type Storer interface {
	Save(ctx context.Context, target Target) error
	Delete(ctx context.Context, target Target) error
	Update(ctx context.Context, t Target) error
	QueryByID(ctx context.Context, id uuid.UUID) (Target, error)
	DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error
	QueryByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]Target, error)
	QueryDue(ctx context.Context, now time.Time) ([]Target, error)
}

type Business struct {
	storer        Storer
	campaignQuery CampaignQuerier
}

func NewBusiness(storer Storer, campaignQuery CampaignQuerier) *Business {
	return &Business{storer: storer, campaignQuery: campaignQuery}
}

func (b *Business) Save(ctx context.Context, tn NewTarget) (Target, error) {
	cmp, err := b.campaignQuery.QueryByID(ctx, tn.CampaignID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrNotFound) {
			return Target{}, fmt.Errorf("save: campaignID[%s]: %w", tn.CampaignID, ErrCampaignNotFound)
		}
		return Target{}, fmt.Errorf("save: %w", err)
	}

	if cmp.Status != campaignbus.Draft {
		return Target{}, fmt.Errorf("save: %w", ErrCampaignNotDraft)
	}

	t := Target{
		ID:          uuid.New(),
		Token:       generateToken(),
		EmployeeID:  tn.EmployeeID,
		CampaignID:  tn.CampaignID,
		Status:      Pending,
		ScheduledAt: nil,
		CreatedAt:   time.Now().UTC(),
	}

	if err := b.storer.Save(ctx, t); err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return Target{}, fmt.Errorf("save: campaignID[%s]: %w", tn.CampaignID, err)
		}
		if errors.Is(err, ErrEmployeeNotFound) {
			return Target{}, fmt.Errorf("save: employeeID[%s]: %w", tn.EmployeeID, err)
		}
		return Target{}, fmt.Errorf("save: %w", err)
	}

	return t, nil
}

func (b *Business) Delete(ctx context.Context, t Target) error {
	if err := b.storer.Delete(ctx, t); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (b *Business) DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error {
	if _, err := b.campaignQuery.QueryByID(ctx, campaignID); err != nil {
		if errors.Is(err, campaignbus.ErrNotFound) {
			return fmt.Errorf("deletebycampaignid: campaignID[%s]: %w", campaignID, ErrCampaignNotFound)
		}
		return fmt.Errorf("deletebycampaignid: %w", err)
	}

	if err := b.storer.DeleteByCampaignID(ctx, campaignID); err != nil {
		return fmt.Errorf("deletebycampaignid: %w", err)
	}
	return nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Target, error) {
	target, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Target{}, fmt.Errorf("querybyid: targetID[%s]: %w", id, err)
	}
	return target, nil
}

func (b *Business) ChangeStatus(ctx context.Context, t Target, s Status) error {
	if !isValidTransition(t.Status, s) {
		return fmt.Errorf("%w: cannot move from %s to %s", ErrInvalidStatusTransition, t.Status, s)
	}

	t.Status = s

	if err := b.storer.Update(ctx, t); err != nil {
		return fmt.Errorf("changestatus: %w", err)
	}
	return nil
}

func (b *Business) QueryDue(ctx context.Context) ([]Target, error) {
	targets, err := b.storer.QueryDue(ctx, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("querydue: %w", err)
	}
	return targets, nil
}

func (b *Business) UpdateSchedule(ctx context.Context, t Target, scheduledAt time.Time) (Target, error) {
	cmp, err := b.campaignQuery.QueryByID(ctx, t.CampaignID)
	if err != nil {
		if errors.Is(err, campaignbus.ErrNotFound) {
			return Target{}, fmt.Errorf("updateschedule: campaignID[%s]: %w", t.CampaignID, ErrCampaignNotFound)
		}
		return Target{}, fmt.Errorf("updateschedule: %w", err)
	}

	if cmp.Status != campaignbus.Draft {
		return Target{}, fmt.Errorf("updateschedule: %w", ErrCampaignNotDraft)
	}

	if t.Status != Pending {
		return Target{}, fmt.Errorf("updateschedule: %w", ErrTargetNotPending)
	}

	if scheduledAt.Before(cmp.DateRange.Range().Start()) || scheduledAt.After(cmp.DateRange.Range().End()) {
		return Target{}, fmt.Errorf("updateschedule: %w", ErrInvalidScheduleWindow)
	}

	scheduledAt = scheduledAt.UTC()
	t.ScheduledAt = &scheduledAt

	if err := b.storer.Update(ctx, t); err != nil {
		return Target{}, fmt.Errorf("updateschedule: %w", err)
	}
	return t, nil
}

func (b *Business) AutoDistribute(ctx context.Context, cmp uuid.UUID) error {
	campaign, err := b.campaignQuery.QueryByID(ctx, cmp)
	if err != nil {
		if errors.Is(err, campaignbus.ErrNotFound) {
			return fmt.Errorf("autodistribute: campaignID[%s]: %w", cmp, ErrCampaignNotFound)
		}
		return fmt.Errorf("autodistribute: %w", err)
	}

	if !campaign.DateRange.Valid() {
		return fmt.Errorf("autodistribute: %w", ErrInvalidScheduleWindow)
	}

	dateRange := campaign.DateRange.Range()
	from := dateRange.Start()
	to := dateRange.End()

	if !to.After(from) {
		return fmt.Errorf("autodistribute: %w", ErrInvalidScheduleWindow)
	}

	if campaign.Status != campaignbus.Draft {
		return fmt.Errorf("autodistribute: %w", ErrCampaignNotDraft)
	}

	targets, err := b.storer.QueryByCampaignID(ctx, campaign.ID)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return fmt.Errorf("autodistribute: campaignID[%s]: %w", campaign.ID, ErrCampaignNotFound)
		}
		return fmt.Errorf("autodistribute: %w", err)
	}

	pending := make([]Target, 0, len(targets))
	for _, target := range targets {
		if target.Status == Pending {
			pending = append(pending, target)
		}
	}

	if len(pending) == 0 {
		return nil
	}

	from = from.UTC()
	to = to.UTC()

	step := to.Sub(from) / time.Duration(len(pending))
	halfStep := step / 2

	for i, target := range pending {
		scheduledAt := from.Add(step*time.Duration(i) + halfStep)
		target.ScheduledAt = &scheduledAt

		if err := b.storer.Update(ctx, target); err != nil {
			return fmt.Errorf("autodistribute: %w", err)
		}
	}

	return nil
}

func generateToken() string {
	return uuid.New().String()
}

func isValidTransition(current, next Status) bool {
	switch current {
	case Pending:
		return next == Sent || next == Failed
	case Sent:
		return next == Opened || next == Clicked || next == Failed
	case Opened:
		return next == Clicked
	case Clicked:
		return next == Submitted
	default:
		return false
	}
}
