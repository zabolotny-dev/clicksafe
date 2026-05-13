package attachmentbus

import "github.com/zabolotny-dev/clicksafe/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.DESC)

const (
	OrderByID         = "a"
	OrderByLabel      = "b"
	OrderByType       = "c"
	OrderByUploadedAt = "d"
)
