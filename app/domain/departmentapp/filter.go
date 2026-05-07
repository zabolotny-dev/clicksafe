package departmentapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
)

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Label   string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("id"),
		Label:   values.Get("label"),
	}

	return filter
}

func parseFilter(qp queryParams) (departmentbus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter departmentbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			fieldErrors.Add("id", err)
		} else {
			filter.ID = &id
		}
	}

	if qp.Label != "" {
		filter.Label = &qp.Label
	}

	if len(fieldErrors) > 0 {
		return departmentbus.QueryFilter{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return filter, nil
}
