package campaignbus

import "github.com/zabolotny-dev/clicksafe/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.DESC)

const (
	OrderByID       = "a"
	OrderByLabel    = "b"
	OrderByStatus   = "c"
	OrderByDateFrom = "d"
	OrderByDateTo   = "e"
)
