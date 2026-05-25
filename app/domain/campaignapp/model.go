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
	ID                 uuid.UUID         `json:"id"`
	Type               string            `json:"type"`
	MessageID          *uuid.UUID        `json:"message_id"`
	LandingID          *uuid.UUID        `json:"landing_id"`
	EducationID        *uuid.UUID        `json:"education_id"`
	MaxEducationTextID uuid.NullUUID     `json:"max_education_text_id"`
	Label              string            `json:"label"`
	Domain             string            `json:"domain"`
	Status             string            `json:"status"`
	DateFrom           *time.Time        `json:"date_from"`
	DateTo             *time.Time        `json:"date_to"`
	Attributes         map[string]string `json:"attributes"`
}

type NewCampaign struct {
	Type               string            `json:"type"`
	MessageID          string            `json:"message_id"`
	LandingID          string            `json:"landing_id"`
	EducationID        string            `json:"education_id"`
	MaxEducationTextID string            `json:"max_education_text_id"`
	Label              string            `json:"label"`
	Domain             string            `json:"domain"`
	DateFrom           time.Time         `json:"date_from"`
	DateTo             time.Time         `json:"date_to"`
	Attributes         map[string]string `json:"attributes"`
}

type UpdateCampaign struct {
	Type               *string            `json:"type"`
	MessageID          *string            `json:"message_id"`
	LandingID          *string            `json:"landing_id"`
	EducationID        *string            `json:"education_id"`
	MaxEducationTextID *string            `json:"max_education_text_id"`
	Label              *string            `json:"label"`
	Domain             *string            `json:"domain"`
	DateFrom           *time.Time         `json:"date_from"`
	DateTo             *time.Time         `json:"date_to"`
	Attributes         *map[string]string `json:"attributes"`
}

func toBusNewCampaign(req NewCampaign) (campaignbus.NewCampaign, error) {
	var fieldErrors errs.FieldErrors

	cmpType := campaignbus.EmailCampaign
	if req.Type != "" {
		parsed, err := campaignbus.ParseCampaignType(req.Type)
		if err != nil {
			fieldErrors.Add("type", err)
		} else {
			cmpType = parsed
		}
	}

	var mID *uuid.UUID
	if req.MessageID != "" {
		id, err := uuid.Parse(req.MessageID)
		if err != nil {
			fieldErrors.Add("message_id", err)
		} else {
			mID = &id
		}
	}

	var lID *uuid.UUID
	if req.LandingID != "" {
		id, err := uuid.Parse(req.LandingID)
		if err != nil {
			fieldErrors.Add("landing_id", err)
		} else {
			lID = &id
		}
	}

	var eID *uuid.UUID
	if req.EducationID != "" {
		id, err := uuid.Parse(req.EducationID)
		if err != nil {
			fieldErrors.Add("education_id", err)
		} else {
			eID = &id
		}
	}

	var maxEducationTextID uuid.NullUUID
	if req.MaxEducationTextID != "" {
		id, err := uuid.Parse(req.MaxEducationTextID)
		if err != nil {
			fieldErrors.Add("max_education_text_id", err)
		} else {
			maxEducationTextID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	label, err := label.Parse(req.Label)
	if err != nil {
		fieldErrors.Add("label", err)
	}

	var dmn domain.Domain
	if req.Domain != "" {
		parsed, err := domain.Parse(req.Domain)
		if err != nil {
			fieldErrors.Add("domain", err)
		} else {
			dmn = parsed
		}
	} else if cmpType == campaignbus.EmailCampaign {
		fieldErrors.Add("domain", errors.New("domain cannot be empty"))
	}

	dater, err := date.ParseNull(req.DateFrom, req.DateTo)
	if err != nil {
		fieldErrors.Add("date_from", err)
		fieldErrors.Add("date_to", err)
	}

	if len(fieldErrors) > 0 {
		return campaignbus.NewCampaign{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return campaignbus.NewCampaign{
		Type:               cmpType,
		MessageID:          mID,
		LandingID:          lID,
		EducationID:        eID,
		MaxEducationTextID: maxEducationTextID,
		Label:              label,
		Domain:             dmn,
		DateRange:          dater,
		Attributes:         req.Attributes,
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
		ID:                 cmp.ID,
		Type:               cmp.Type.String(),
		MessageID:          cmp.MessageID,
		LandingID:          cmp.LandingID,
		EducationID:        cmp.EducationID,
		MaxEducationTextID: cmp.MaxEducationTextID,
		Label:              cmp.Label.String(),
		Domain:             cmp.Domain.String(),
		Status:             cmp.Status.String(),
		DateFrom:           dateFrom,
		DateTo:             dateTo,
		Attributes:         cmp.Attributes,
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

	var cmpType *campaignbus.CampaignType
	if req.Type != nil {
		parsed, err := campaignbus.ParseCampaignType(*req.Type)
		if err != nil {
			fieldErrors.Add("type", err)
		} else {
			cmpType = &parsed
		}
	}

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

	var eID *uuid.UUID
	if req.EducationID != nil {
		id, err := uuid.Parse(*req.EducationID)
		if err != nil {
			fieldErrors.Add("education_id", err)
		} else {
			eID = &id
		}
	}

	var maxEducationTextID *uuid.NullUUID
	if req.MaxEducationTextID != nil {
		maxEducationTextID = &uuid.NullUUID{}
		if id := *req.MaxEducationTextID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				fieldErrors.Add("max_education_text_id", err)
			}
			maxEducationTextID.UUID = parsed
			maxEducationTextID.Valid = err == nil
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
		dmn = &domain.Domain{}
		if *req.Domain != "" {
			parsed, err := domain.Parse(*req.Domain)
			if err != nil {
				fieldErrors.Add("domain", err)
			}
			dmn = &parsed
		}
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
		Type:               cmpType,
		MessageID:          mID,
		LandingID:          lID,
		EducationID:        eID,
		MaxEducationTextID: maxEducationTextID,
		Label:              lbl,
		Domain:             dmn,
		DateRange:          dater,
		Attributes:         req.Attributes,
	}, nil
}
