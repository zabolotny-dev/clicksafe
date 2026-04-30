package eventapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/types/event"
)

type Event struct {
	CampaignID string `json:"campaign_id"`
	EmployeeID string `json:"employee_id"`
	Type       string `json:"type"`
}

func toBusEvent(e Event) (eventbus.Event, error) {
	var errors errs.FieldErrors

	campID, err := uuid.Parse(e.CampaignID)
	if err != nil {
		errors.Add("campaign_id", err)
	}

	employeeID, err := uuid.Parse(e.EmployeeID)
	if err != nil {
		errors.Add("employee_id", err)
	}

	evType, err := event.Parse(e.Type)
	if err != nil {
		errors.Add("type", err)
	}

	if len(errors) > 0 {
		return eventbus.Event{}, errors.ToError()
	}

	return eventbus.Event{
		ID:         uuid.New(),
		CampaignID: campID,
		EmployeeID: employeeID,
		Type:       evType,
	}, nil
}
