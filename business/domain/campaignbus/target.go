package campaignbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Target struct {
	ID          uuid.UUID
	Token       string
	EmployeeID  uuid.UUID
	CampaignID  uuid.UUID
	Status      TargetStatus
	ScheduledAt *time.Time
	CreatedAt   time.Time
}

type NewTarget struct {
	EmployeeID uuid.UUID
	CampaignID uuid.UUID
}

type TargetStorer interface {
	Save(ctx context.Context, target Target) error
	Delete(ctx context.Context, target Target) error
	Update(ctx context.Context, t Target) error
	UpdateMany(ctx context.Context, t []Target) error
	QueryByID(ctx context.Context, id uuid.UUID) (Target, error)
	QueryByToken(ctx context.Context, token string) (Target, error)
	DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error
	QueryDue(ctx context.Context, now time.Time) ([]Target, error)
	Query(ctx context.Context, filter TargetQueryFilter) ([]Target, error)
	Count(ctx context.Context, filter TargetQueryFilter) (int, error)
}

func (b *TargetBusiness) Save(ctx context.Context, tn NewTarget) (Target, error) {
	cmp, err := b.campaignStorer.QueryByID(ctx, tn.CampaignID)
	if err != nil {
		return Target{}, fmt.Errorf("save: campaignID[%s]: %w", tn.CampaignID, err)
	}

	if cmp.Status != Draft && cmp.Status != Paused {
		return Target{}, fmt.Errorf("save: %w: cannot add targets in %s status", ErrCampaignLocked, cmp.Status)
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

	if err := b.targetStorer.Save(ctx, t); err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return Target{}, fmt.Errorf("save: campaignID[%s]: %w", tn.CampaignID, err)
		}
		if errors.Is(err, ErrEmployeeNotFound) {
			return Target{}, fmt.Errorf("save: employeeID[%s]: %w", tn.EmployeeID, err)
		}
		if errors.Is(err, ErrTargetAlreadyExists) {
			return Target{}, fmt.Errorf("save: employeeID[%s]: %w", tn.EmployeeID, err)
		}
		return Target{}, fmt.Errorf("save: %w", err)
	}

	return t, nil
}

func (b *TargetBusiness) Delete(ctx context.Context, t Target) error {
	if err := b.targetStorer.Delete(ctx, t); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (b *TargetBusiness) DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error {
	if _, err := b.campaignStorer.QueryByID(ctx, campaignID); err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return fmt.Errorf("deletebycampaignid: campaignID[%s]: %w", campaignID, ErrCampaignNotFound)
		}
		return fmt.Errorf("deletebycampaignid: %w", err)
	}

	if err := b.targetStorer.DeleteByCampaignID(ctx, campaignID); err != nil {
		return fmt.Errorf("deletebycampaignid: %w", err)
	}
	return nil
}

func (b *TargetBusiness) QueryByID(ctx context.Context, id uuid.UUID) (Target, error) {
	target, err := b.targetStorer.QueryByID(ctx, id)
	if err != nil {
		return Target{}, fmt.Errorf("querybyid: targetID[%s]: %w", id, err)
	}
	return target, nil
}

func (b *TargetBusiness) QueryCampaignByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	campaign, err := b.campaignStorer.QueryByID(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("querycampaignbyid: campaignID[%s]: %w", id, err)
	}
	return campaign, nil
}

func (b *TargetBusiness) QueryByToken(ctx context.Context, token string) (Target, error) {
	target, err := b.targetStorer.QueryByToken(ctx, token)
	if err != nil {
		return Target{}, fmt.Errorf("querybytoken: token[%s]: %w", token, err)
	}
	return target, nil
}

func (b *TargetBusiness) ChangeStatus(ctx context.Context, t Target, s TargetStatus) error {
	if !isValidTargetTransition(t.Status, s) {
		return fmt.Errorf("changestatus: %w: cannot move from %s to %s", ErrInvalidStatusTransition, t.Status, s)
	}

	t.Status = s

	if err := b.targetStorer.Update(ctx, t); err != nil {
		return fmt.Errorf("changestatus: %w", err)
	}
	return nil
}

func (b *TargetBusiness) QueryDue(ctx context.Context) ([]Target, error) {
	targets, err := b.targetStorer.QueryDue(ctx, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("querydue: %w", err)
	}
	return targets, nil
}

func (b *TargetBusiness) UpdateSchedule(ctx context.Context, t Target, scheduledAt time.Time) (Target, error) {
	cmp, err := b.campaignStorer.QueryByID(ctx, t.CampaignID)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return Target{}, fmt.Errorf("updateschedule: campaignID[%s]: %w", t.CampaignID, ErrCampaignNotFound)
		}
		return Target{}, fmt.Errorf("updateschedule: %w", err)
	}

	if cmp.Status != Draft && cmp.Status != Paused {
		return Target{}, fmt.Errorf("updateschedule: %w: cannot change schedule in %s campaign status", ErrTargetLocked, cmp.Status)
	}

	if t.Status != Pending {
		return Target{}, fmt.Errorf("updateschedule: %w: cannot change schedule in %s target status", ErrTargetLocked, t.Status)
	}

	if !cmp.DateRange.Range().Contains(scheduledAt) {
		return Target{}, fmt.Errorf("updateschedule: %w", ErrInvalidScheduleWindow)
	}

	scheduledAt = scheduledAt.UTC()
	t.ScheduledAt = &scheduledAt

	if err := b.targetStorer.Update(ctx, t); err != nil {
		return Target{}, fmt.Errorf("updateschedule: %w", err)
	}
	return t, nil
}

func (b *TargetBusiness) AutoDistribute(ctx context.Context, cmp uuid.UUID) error {
	campaign, err := b.campaignStorer.QueryByID(ctx, cmp)
	if err != nil {
		return fmt.Errorf("autodistribute: campaignID[%s]: %w", cmp, err)
	}

	if !campaign.DateRange.Valid() {
		return fmt.Errorf("autodistribute: %w", ErrInvalidScheduleWindow)
	}

	if campaign.Status != Draft && campaign.Status != Paused {
		return fmt.Errorf("autodistribute: %w: cannot change schedule in %s campaign status", ErrTargetLocked, campaign.Status)
	}

	targets, err := b.targetStorer.Query(ctx, TargetQueryFilter{
		CampaignID: &campaign.ID,
		Status:     &Pending,
	})

	if err != nil {
		return fmt.Errorf("autodistribute: %w", err)
	}

	if len(targets) == 0 {
		return nil
	}

	dateRange := campaign.DateRange.Range()
	from := dateRange.Start().UTC()
	to := dateRange.End().UTC()

	step := to.Sub(from) / time.Duration(len(targets))
	halfStep := step / 2

	for i := range targets {
		scheduledAt := from.Add(step*time.Duration(i) + halfStep)
		targets[i].ScheduledAt = &scheduledAt
	}

	if err := b.targetStorer.UpdateMany(ctx, targets); err != nil {
		return fmt.Errorf("autodistribute: %w", err)
	}

	return nil
}

func (b *TargetBusiness) PhishingURL(ctx context.Context, id uuid.UUID) (string, error) {
	t, err := b.targetStorer.QueryByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("phishingurl: targetID[%s]: %w", id, err)
	}

	cmp, err := b.campaignStorer.QueryByID(ctx, t.CampaignID)
	if err != nil {
		return "", fmt.Errorf("phishingurl: campaignID[%s]: %w", t.CampaignID, err)
	}

	if cmp.Domain.IsEmpty() {
		return "", fmt.Errorf("phishingurl: campaignID[%s]: %w", cmp.ID, ErrDomainRequired)
	}

	return cmp.Domain.String() + "/" + t.Token, nil
}

func generateToken() string {
	return uuid.New().String()
}

func isValidTargetTransition(current, next TargetStatus) bool {
	switch current {
	case Pending:
		return next == Sent || next == Failed || next == Clicked
	case Sent:
		return next == Opened || next == Clicked || next == Failed || next == Replied
	case Opened:
		return next == Clicked || next == Replied
	case Clicked:
		return next == Submitted || next == Replied
	default:
		return false
	}
}
