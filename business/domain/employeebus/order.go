package employeebus

import "github.com/zabolotny-dev/clicksafe/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.DESC)

const (
	OrderByID        = "a"
	OrderByFirstName = "b"
	OrderByLastName  = "c"
	OrderByEmail     = "d"
	OrderByPhone     = "e"
)
