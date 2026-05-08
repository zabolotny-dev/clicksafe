package vtargetapp

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
)

type queryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	CampaignID string
	EmployeeID string
	Status     string
	FullName   string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	return queryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("target_id"),
		CampaignID: values.Get("campaign_id"),
		EmployeeID: values.Get("employee_id"),
		Status:     values.Get("status"),
		FullName:   values.Get("full_name"),
	}
}

func parseFilter(qp queryParams) (vtargetbus.Filter, error) {
	var fieldErrors errs.FieldErrors
	var filter vtargetbus.Filter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			fieldErrors.Add("target_id", err)
		} else {
			filter.ID = &id
		}
	}

	if qp.CampaignID != "" {
		id, err := uuid.Parse(qp.CampaignID)
		if err != nil {
			fieldErrors.Add("campaign_id", err)
		} else {
			filter.CampaignID = &id
		}
	}

	if qp.EmployeeID != "" {
		id, err := uuid.Parse(qp.EmployeeID)
		if err != nil {
			fieldErrors.Add("employee_id", err)
		} else {
			filter.EmployeeID = &id
		}
	}

	if qp.Status != "" {
		status, err := campaignbus.ParseTargetStatus(qp.Status)
		if err != nil {
			fieldErrors.Add("status", err)
		} else {
			statusValue := status.String()
			filter.Status = &statusValue
		}
	}

	if qp.FullName != "" {
		filter.FullName = &qp.FullName
	}

	if len(fieldErrors) > 0 {
		return vtargetbus.Filter{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return filter, nil
}
