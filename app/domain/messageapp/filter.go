package messageapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

type queryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	Label     string
	FromEmail string
	FromName  string
	Subject   string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		Label:     values.Get("label"),
		FromEmail: values.Get("from_email"),
		FromName:  values.Get("from_name"),
		Subject:   values.Get("subject"),
	}

	return filter
}

func parseFilter(qp queryParams) (messagebus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter messagebus.QueryFilter

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

	if qp.FromEmail != "" {
		filter.FromEmail = &qp.FromEmail
	}

	if qp.FromName != "" {
		filter.FromName = &qp.FromName
	}

	if qp.Subject != "" {
		filter.Subject = &qp.Subject
	}

	if len(fieldErrors) > 0 {
		return messagebus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
