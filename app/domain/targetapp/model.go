package targetapp

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus"
)

type Target struct {
	ID          uuid.UUID  `json:"id"`
	Token       string     `json:"token"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	CampaignID  uuid.UUID  `json:"campaign_id"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type NewTarget struct {
	EmployeeID string `json:"employee_id"`
	CampaignID string `json:"campaign_id"`
}

type UpdateSchedule struct {
	ScheduledAt time.Time `json:"scheduled_at"`
}

type AutoDistribute struct {
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
}

func toBusNewTarget(req NewTarget) (targetbus.NewTarget, error) {
	var errors errs.FieldErrors

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		errors.Add("employee_id", err)
	}

	campaignID, err := uuid.Parse(req.CampaignID)
	if err != nil {
		errors.Add("campaign_id", err)
	}

	if len(errors) > 0 {
		return targetbus.NewTarget{}, errors.ToError()
	}

	return targetbus.NewTarget{
		EmployeeID: employeeID,
		CampaignID: campaignID,
	}, nil
}

func toBusUpdateSchedule(req UpdateSchedule) (time.Time, error) {
	if req.ScheduledAt.IsZero() {
		return time.Time{}, errs.NewFieldErrors("scheduled_at", errors.New("scheduled_at is required"))
	}

	return req.ScheduledAt, nil
}

func toAppTarget(t targetbus.Target) Target {
	return Target{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      t.Status.String(),
		ScheduledAt: t.ScheduledAt,
		CreatedAt:   t.CreatedAt,
	}
}
