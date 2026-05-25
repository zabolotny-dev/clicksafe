package campaigndb

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/campaigndb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBCampaign(c campaignbus.Campaign) (sqlc.Campaign, error) {
	attributes, err := json.Marshal(c.Attributes)
	if err != nil {
		return sqlc.Campaign{}, err
	}

	dateFrom, dateTo := c.DateRange.ToSQLNullTimestamptz()

	return sqlc.Campaign{
		ID:                 c.ID,
		Type:               c.Type.String(),
		MessageID:          c.MessageID,
		LandingID:          c.LandingID,
		EducationID:        c.EducationID,
		MaxEducationTextID: nullUUIDPtr(c.MaxEducationTextID),
		Label:              c.Label.String(),
		Domain:             c.Domain.String(),
		Status:             c.Status.String(),
		DateFrom:           dateFrom,
		DateTo:             dateTo,
		Attributes:         attributes,
	}, nil
}

func toBusCampaign(c sqlc.Campaign) (campaignbus.Campaign, error) {
	var attributes map[string]string
	if len(c.Attributes) > 0 {
		if err := json.Unmarshal(c.Attributes, &attributes); err != nil {
			return campaignbus.Campaign{}, err
		}
	}

	campaignLabel, err := label.Parse(c.Label)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	var campaignDomain domain.Domain
	if c.Domain != "" {
		parsed, err := domain.Parse(c.Domain)
		if err != nil {
			return campaignbus.Campaign{}, err
		}
		campaignDomain = parsed
	}

	status, err := campaignbus.ParseCampaignStatus(c.Status)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	cmpType, err := campaignbus.ParseCampaignType(c.Type)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	dateRange, err := date.ParseSQLNullTimestamptz(c.DateFrom, c.DateTo)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	return campaignbus.Campaign{
		ID:                 c.ID,
		Type:               cmpType,
		MessageID:          c.MessageID,
		LandingID:          c.LandingID,
		EducationID:        c.EducationID,
		MaxEducationTextID: ptrNullUUID(c.MaxEducationTextID),
		Label:              campaignLabel,
		Domain:             campaignDomain,
		Status:             status,
		DateRange:          dateRange,
		Attributes:         attributes,
	}, nil
}

func toBusCampaigns(campaigns []sqlc.Campaign) ([]campaignbus.Campaign, error) {
	busCampaigns := make([]campaignbus.Campaign, len(campaigns))

	for i, c := range campaigns {
		var err error
		busCampaigns[i], err = toBusCampaign(c)
		if err != nil {
			return nil, err
		}
	}

	return busCampaigns, nil
}

func nullUUIDPtr(id uuid.NullUUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	return &id.UUID
}

func ptrNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}
