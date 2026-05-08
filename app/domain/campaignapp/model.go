package campaignapp

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Campaign struct {
	ID         uuid.UUID         `json:"id"`
	MessageID  *uuid.UUID        `json:"message_id"`
	LandingID  *uuid.UUID        `json:"landing_id"`
	Label      string            `json:"label"`
	Domain     string            `json:"domain"`
	Status     string            `json:"status"`
	DateFrom   *time.Time        `json:"date_from"`
	DateTo     *time.Time        `json:"date_to"`
	Attributes map[string]string `json:"attributes"`
}

type NewCampaign struct {
	MessageID  string            `json:"message_id"`
	LandingID  string            `json:"landing_id"`
	Label      string            `json:"label"`
	Domain     string            `json:"domain"`
	DateFrom   time.Time         `json:"date_from"`
	DateTo     time.Time         `json:"date_to"`
	Attributes map[string]string `json:"attributes"`
}

type UpdateCampaign struct {
	MessageID  *string            `json:"message_id"`
	LandingID  *string            `json:"landing_id"`
	Label      *string            `json:"label"`
	Domain     *string            `json:"domain"`
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

	var lID *uuid.UUID
	if req.LandingID != "" {
		id, err := uuid.Parse(req.LandingID)
		if err != nil {
			errors.Add("landing_id", err)
		} else {
			lID = &id
		}
	}

	label, err := label.Parse(req.Label)
	if err != nil {
		errors.Add("label", err)
	}

	domain, err := domain.Parse(req.Domain)
	if err != nil {
		errors.Add("domain", err)
	}

	dater, err := date.ParseNull(req.DateFrom, req.DateTo)
	if err != nil {
		errors.Add("date_from", err)
		errors.Add("date_to", err)
	}

	if len(errors) > 0 {
		return campaignbus.NewCampaign{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return campaignbus.NewCampaign{
		MessageID:  mID,
		LandingID:  lID,
		Label:      label,
		Domain:     domain,
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
		LandingID:  cmp.LandingID,
		Label:      cmp.Label.String(),
		Domain:     cmp.Domain.String(),
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
	var fieldErrors errs.FieldErrors

	var mID *uuid.UUID
	if req.MessageID != nil {
		id, err := uuid.Parse(*req.MessageID)
		if err != nil {
			fieldErrors.Add("message_id", err)
		} else {
			mID = &id
		}
	}

	var lID *uuid.UUID
	if req.LandingID != nil {
		id, err := uuid.Parse(*req.LandingID)
		if err != nil {
			fieldErrors.Add("landing_id", err)
		} else {
			lID = &id
		}
	}

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			fieldErrors.Add("label", err)
		}
		lbl = &parsed
	}

	var dmn *domain.Domain
	if req.Domain != nil {
		parsed, err := domain.Parse(*req.Domain)
		if err != nil {
			fieldErrors.Add("domain", err)
		}
		dmn = &parsed
	}

	var dater *date.Null
	switch {
	case req.DateFrom == nil && req.DateTo == nil:
	case req.DateFrom == nil:
		fieldErrors.Add("date_from", errors.New("date_from is required when date_to is provided"))
	case req.DateTo == nil:
		fieldErrors.Add("date_to", errors.New("date_to is required when date_from is provided"))
	default:
		parsed, err := date.ParseNull(*req.DateFrom, *req.DateTo)
		if err != nil {
			fieldErrors.Add("date_from", err)
			fieldErrors.Add("date_to", err)
		} else {
			dater = &parsed
		}
	}

	if len(fieldErrors) > 0 {
		return campaignbus.UpdateCampaign{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return campaignbus.UpdateCampaign{
		MessageID:  mID,
		LandingID:  lID,
		Label:      lbl,
		Domain:     dmn,
		DateRange:  dater,
		Attributes: req.Attributes,
	}, nil
}
