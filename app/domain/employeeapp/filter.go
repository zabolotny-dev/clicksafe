package employeeapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
)

type queryParams struct {
	Page         string
	Rows         string
	OrderBy      string
	ID           string
	DepartmentID string
	Fullname     string
	Email        string
	Phone        string
}

func parseQueryParams(c *echo.Context) (queryParams, error) {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:         values.Get("page"),
		Rows:         values.Get("rows"),
		OrderBy:      values.Get("orderBy"),
		ID:           values.Get("employee_id"),
		DepartmentID: values.Get("department_id"),
		Fullname:     values.Get("full_name"),
		Email:        values.Get("email"),
		Phone:        values.Get("phone"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (employeebus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter employeebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			fieldErrors.Add("id", err)
		} else {
			filter.ID = &id
		}
	}

	if qp.DepartmentID != "" {
		id, err := uuid.Parse(qp.DepartmentID)
		if err != nil {
			fieldErrors.Add("department_id", err)
		} else {
			filter.DepartmentID = &id
		}
	}

	if qp.Fullname != "" {
		filter.FullName = &qp.Fullname
	}

	if qp.Email != "" {
		filter.Email = &qp.Email
	}

	if qp.Phone != "" {
		filter.Phone = &qp.Phone
	}

	if len(fieldErrors) > 0 {
		return employeebus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
