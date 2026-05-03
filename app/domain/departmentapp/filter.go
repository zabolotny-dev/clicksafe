package departmentapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Label   string
}

func parseQueryParams(c *echo.Context) (queryParams, error) {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("department_id"),
		Label:   values.Get("label"),
	}

	return filter, nil
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
		lbl, err := label.Parse(qp.Label)
		if err != nil {
			fieldErrors.Add("label", err)
		} else {
			filter.Label = &lbl
		}
	}

	if len(fieldErrors) > 0 {
		return departmentbus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
