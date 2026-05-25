package maxaccountapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
)

type queryParams struct {
	Page  string
	Rows  string
	Label string
	Phone string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	return queryParams{
		Page:  values.Get("page"),
		Rows:  values.Get("rows"),
		Label: values.Get("label"),
		Phone: values.Get("phone"),
	}
}

func parseFilter(qp queryParams) (maxaccountbus.QueryFilter, error) {
	var filter maxaccountbus.QueryFilter

	if qp.Label != "" {
		filter.Label = &qp.Label
	}
	if qp.Phone != "" {
		filter.Phone = &qp.Phone
	}

	return filter, nil
}
