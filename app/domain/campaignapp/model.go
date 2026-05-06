package campaignapp

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Campaign struct {
	ID         uuid.UUID         `json:"id"`
	MessageID  *uuid.UUID        `json:"message_id"`
	Label      string            `json:"label"`
	Status     string            `json:"status"`
	DateFrom   *time.Time        `json:"date_from"`
	DateTo     *time.Time        `json:"date_to"`
	Attributes map[string]string `json:"attributes"`
}

type NewCampaign struct {
	MessageID  string            `json:"message_id"`
	Label      string            `json:"label"`
	DateFrom   time.Time         `json:"date_from"`
	DateTo     time.Time         `json:"date_to"`
	Attributes map[string]string `json:"attributes"`
}

type UpdateCampaign struct {
	MessageID  *string            `json:"message_id"`
	Label      *string            `json:"label"`
	DateFrom   *time.Time         `json:"date_from"`
	DateTo     *time.Time         `json:"date_to"`
	Attributes *map[string]string `json:"attributes"`
}

func toBusNewCampaign(req NewCampaign) (campaignbus.NewCampaign, error) {
	var errors errs.FieldErrors

	var mID *uuid.UUID
	if req.MessageID != "" {
		id, err := uuid.Parse(req.MessageID)
		if err != nil {
			errors.Add("message_id", err)
		} else {
			mID = &id
		}
	}

	label, err := label.Parse(req.Label)
	if err != nil {
		errors.Add("label", err)
	}

	dater, err := date.ParseNull(req.DateFrom, req.DateTo)
	if err != nil {
		errors.Add("date_from", err)
		errors.Add("date_to", err)
	}

	if len(errors) > 0 {
		return campaignbus.NewCampaign{}, errors.ToError()
	}

	return campaignbus.NewCampaign{
		MessageID:  mID,
		Label:      label,
		DateRange:  dater,
		Attributes: req.Attributes,
	}, nil
}

func toAppCampaign(cmp campaignbus.Campaign) Campaign {
	var dateFrom, dateTo *time.Time
	if cmp.DateRange.Valid() {
		rng := cmp.DateRange.Range()
		start := rng.Start()
		end := rng.End()
		dateFrom = &start
		dateTo = &end
	}

	return Campaign{
		ID:         cmp.ID,
		MessageID:  cmp.MessageID,
		Label:      cmp.Label.String(),
		Status:     cmp.Status.String(),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Attributes: cmp.Attributes,
	}
}

func toAppCampaigns(campaigns []campaignbus.Campaign) []Campaign {
	result := make([]Campaign, len(campaigns))
	for i, c := range campaigns {
		result[i] = toAppCampaign(c)
	}
	return result
}

func toBusUpdateCampaign(req UpdateCampaign) (campaignbus.UpdateCampaign, error) {
	var errs errs.FieldErrors

	var mID *uuid.UUID
	if req.MessageID != nil {
		id, err := uuid.Parse(*req.MessageID)
		if err != nil {
			errs.Add("message_id", err)
		} else {
			mID = &id
		}
	}

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			errs.Add("label", err)
		}
		lbl = &parsed
	}

	var dater *date.Null
	switch {
	case req.DateFrom == nil && req.DateTo == nil:
	case req.DateFrom == nil:
		errs.Add("date_from", errors.New("date_from is required when date_to is provided"))
	case req.DateTo == nil:
		errs.Add("date_to", errors.New("date_to is required when date_from is provided"))
	default:
		parsed, err := date.ParseNull(*req.DateFrom, *req.DateTo)
		if err != nil {
			errs.Add("date_from", err)
			errs.Add("date_to", err)
		} else {
			dater = &parsed
		}
	}

	if len(errs) > 0 {
		return campaignbus.UpdateCampaign{}, errs.ToError()
	}

	return campaignbus.UpdateCampaign{
		MessageID:  mID,
		Label:      lbl,
		DateRange:  dater,
		Attributes: req.Attributes,
	}, nil
}
