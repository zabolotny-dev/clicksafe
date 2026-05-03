package messagebus

import "github.com/zabolotny-dev/clicksafe/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByLabel, order.DESC)

const (
	OrderByID      = "a"
	OrderByLabel   = "b"
	OrderByEmail   = "c"
	OrderByName    = "d"
	OrderBySubject = "e"
)
