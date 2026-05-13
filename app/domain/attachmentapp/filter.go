package attachmentapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
)

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Label   string
	Type    string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("id"),
		Label:   values.Get("label"),
		Type:    values.Get("type"),
	}

	return filter
}

func parseFilter(qp queryParams) (attachmentbus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter attachmentbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			fieldErrors.Add("id", err)
		}
		filter.ID = &id
	}

	if qp.Label != "" {
		filter.Label = &qp.Label
	}

	if qp.Type != "" {
		tp, err := attachmentbus.Parse(qp.Type)
		if err != nil {
			fieldErrors.Add("type", err)
		}
		filter.Type = &tp
	}

	if len(fieldErrors) > 0 {
		return attachmentbus.QueryFilter{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return filter, nil
}
