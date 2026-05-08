package landingapp

import "github.com/zabolotny-dev/clicksafe/business/domain/landingbus"

var orderByFields = map[string]string{
	"landing_id": landingbus.OrderByID,
	"label":      landingbus.OrderByLabel,
}
