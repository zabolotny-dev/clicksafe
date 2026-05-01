package departmentbus

import "github.com/zabolotny-dev/clicksafe/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.DESC)

const (
	OrderByID   = "a"
	OrderByName = "b"
)
