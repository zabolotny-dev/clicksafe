package campaignapp

import "github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"

var orderByFields = map[string]string{
	"campaign_id": campaignbus.OrderByID,
	"label":       campaignbus.OrderByLabel,
	"status":      campaignbus.OrderByStatus,
	"date_from":   campaignbus.OrderByDateFrom,
	"date_to":     campaignbus.OrderByDateTo,
}
