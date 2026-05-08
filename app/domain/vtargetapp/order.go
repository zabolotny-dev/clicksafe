package vtargetapp

import "github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"

var orderByFields = map[string]string{
	"target_id":  vtargetbus.OrderByID,
	"first_name": vtargetbus.OrderByFirstName,
	"last_name":  vtargetbus.OrderByLastName,
	"status":     vtargetbus.OrderByStatus,
}
