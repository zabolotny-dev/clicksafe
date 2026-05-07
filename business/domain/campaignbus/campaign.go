package campaignbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Campaign struct {
	ID         uuid.UUID
	MessageID  *uuid.UUID
	Label      label.Label
	Status     CampaignStatus
	DateRange  date.Null
	Attributes map[string]string
}

type NewCampaign struct {
	MessageID  *uuid.UUID
	Label      label.Label
	DateRange  date.Null
	Attributes map[string]string
}

type UpdateCampaign struct {
	MessageID  *uuid.UUID
	Label      *label.Label
	DateRange  *date.Null
	Attributes *map[string]string
}

type CampaignStorer interface {
	Save(ctx context.Context, campaign Campaign) error
	Update(ctx context.Context, campaign Campaign) error
	QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error)
	Query(ctx context.Context, filter CampaignQueryFilter, orderBy order.By, page page.Page) ([]Campaign, error)
	Delete(ctx context.Context, campaign Campaign) error
	Count(ctx context.Context, filter CampaignQueryFilter) (int, error)
	QueryExpired(ctx context.Context) ([]Campaign, error)
}

func (b *CampaignBusiness) Save(ctx context.Context, campaign NewCampaign) (Campaign, error) {
	cmp := Campaign{
		ID:         uuid.New(),
		MessageID:  campaign.MessageID,
		Label:      campaign.Label,
		Status:     Draft,
		DateRange:  campaign.DateRange,
		Attributes: campaign.Attributes,
	}

	if err := b.campaignStorer.Save(ctx, cmp); err != nil {
		if errors.Is(err, ErrMessageNotFound) && campaign.MessageID != nil {
			return Campaign{}, fmt.Errorf("save: messageID[%s]: %w", *campaign.MessageID, err)
		}
		return Campaign{}, fmt.Errorf("save: %w", err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) Update(ctx context.Context, cmp Campaign, upd UpdateCampaign) (Campaign, error) {
	if upd.MessageID != nil {
		cmp.MessageID = upd.MessageID
	}

	if upd.Label != nil {
		cmp.Label = *upd.Label
	}

	if upd.DateRange != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change date in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.DateRange = *upd.DateRange
	}

	if upd.Attributes != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change attributes in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.Attributes = *upd.Attributes
	}

	if err := b.campaignStorer.Update(ctx, cmp); err != nil {
		if errors.Is(err, ErrMessageNotFound) && upd.MessageID != nil {
			return Campaign{}, fmt.Errorf("update: messageID[%s]: %w", *upd.MessageID, err)
		}
		return Campaign{}, fmt.Errorf("update: %w", err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	cmp, err := b.campaignStorer.QueryByID(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("query: campaignID[%s]: %w", id, err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) Query(ctx context.Context, filter CampaignQueryFilter, orderBy order.By, page page.Page) ([]Campaign, error) {
	camps, err := b.campaignStorer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return camps, nil
}

func (b *CampaignBusiness) Delete(ctx context.Context, campaign Campaign) error {
	if err := b.campaignStorer.Delete(ctx, campaign); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// TODO решить как там где проверять, что сообщение имеет html файл
// не хватает проверки что хватает всех полей у таргетов для рендера шаблонов
func (b *CampaignBusiness) Start(ctx context.Context, campaign Campaign) (Campaign, error) {
	if !isValidCampaignTransition(campaign.Status, Active) {
		return Campaign{}, fmt.Errorf("start: %w: cannot move from %s to %s", ErrInvalidStatusTransition, campaign.Status, Active)
	}

	if campaign.MessageID == nil {
		return Campaign{}, fmt.Errorf("start: %w", ErrMessageRequired)
	}

	if !campaign.DateRange.Valid() {
		return Campaign{}, fmt.Errorf("start: %w", ErrDateRangeRequired)
	}

	if !campaign.DateRange.Range().End().After(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("start: %w", ErrDateRangeExpired)
	}

	targetCount, err := b.targetStorer.Count(ctx, TargetQueryFilter{CampaignID: &campaign.ID})
	if err != nil {
		return Campaign{}, fmt.Errorf("start: count targets: %w", err)
	}

	if targetCount == 0 {
		return Campaign{}, fmt.Errorf("start: %w", ErrTargetsRequired)
	}

	hasSchedule := false
	noScheduleTargets, err := b.targetStorer.Query(ctx, TargetQueryFilter{
		CampaignID:  &campaign.ID,
		HasSchedule: &hasSchedule,
	})
	if err != nil {
		return Campaign{}, fmt.Errorf("start: query targets: %w", err)
	}

	if len(noScheduleTargets) > 0 {
		ids := make([]uuid.UUID, len(noScheduleTargets))
		for i, t := range noScheduleTargets {
			ids[i] = t.ID
		}
		return Campaign{}, fmt.Errorf("start: %w", &ErrUnscheduledTargets{TargetIDs: ids})
	}

	return b.changeStatus(ctx, campaign, Active, "start")
}

func (b *CampaignBusiness) Pause(ctx context.Context, campaign Campaign) (Campaign, error) {
	return b.changeStatus(ctx, campaign, Paused, "pause")
}

func (b *CampaignBusiness) Cancel(ctx context.Context, campaign Campaign) (Campaign, error) {
	return b.changeStatus(ctx, campaign, Canceled, "cancel")
}

func (b *CampaignBusiness) Count(ctx context.Context, filter CampaignQueryFilter) (int, error) {
	count, err := b.campaignStorer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func (b *CampaignBusiness) CompleteExpired(ctx context.Context) []error {
	var errs []error
	campaigns, err := b.campaignStorer.QueryExpired(ctx)

	if err != nil {
		errs = append(errs, fmt.Errorf("completeexpired: %w", err))
		return errs
	}

	if len(campaigns) == 0 {
		return nil
	}

	var readyCmp []Campaign
	for _, campaign := range campaigns {
		targetCount, err := b.targetStorer.Count(ctx, TargetQueryFilter{CampaignID: &campaign.ID, Status: &Pending})
		if err != nil {
			errs = append(errs, fmt.Errorf("completeexpired: campaignID[%s]: %w", campaign.ID, err))
			continue
		}
		if targetCount != 0 {
			continue
		}
		readyCmp = append(readyCmp, campaign)
	}

	for _, campaign := range readyCmp {
		if _, err := b.changeStatus(ctx, campaign, Completed, "completeexpired"); err != nil {
			errs = append(errs, fmt.Errorf("completeexpired: campaignID[%s]: %w", campaign.ID, err))
		}
	}

	return errs
}

func (b *CampaignBusiness) changeStatus(ctx context.Context, campaign Campaign, status CampaignStatus, op string) (Campaign, error) {
	if !isValidCampaignTransition(campaign.Status, status) {
		return Campaign{}, fmt.Errorf("%s: %w: cannot move from %s to %s", op, ErrInvalidStatusTransition, campaign.Status, status)
	}

	campaign.Status = status

	if err := b.campaignStorer.Update(ctx, campaign); err != nil {
		return Campaign{}, fmt.Errorf("%s: %w", op, err)
	}

	return campaign, nil
}

func isValidCampaignTransition(current, next CampaignStatus) bool {
	switch current {
	case Draft:
		return next == Active || next == Canceled
	case Active:
		return next == Paused || next == Canceled || next == Completed
	case Paused:
		return next == Active || next == Canceled
	default:
		return false
	}
}
