package campaigndb

import (
	"encoding/json"

	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/campaigndb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBCampaign(c campaignbus.Campaign) (sqlc.Campaign, error) {
	attributes, err := json.Marshal(c.Attributes)
	if err != nil {
		return sqlc.Campaign{}, err
	}

	dateFrom, dateTo := c.DateRange.ToSQLNullTimestamptz()

	return sqlc.Campaign{
		ID:         c.ID,
		MessageID:  c.MessageID,
		Label:      c.Label.String(),
		Status:     c.Status.String(),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Attributes: attributes,
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

	status, err := campaignbus.Parse(c.Status)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	dateRange, err := date.ParseSQLNullTimestamptz(c.DateFrom, c.DateTo)
	if err != nil {
		return campaignbus.Campaign{}, err
	}

	return campaignbus.Campaign{
		ID:         c.ID,
		MessageID:  c.MessageID,
		Label:      campaignLabel,
		Status:     status,
		DateRange:  dateRange,
		Attributes: attributes,
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
