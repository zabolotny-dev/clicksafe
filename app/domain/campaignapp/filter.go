package campaignapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type queryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Type     string
	Label    string
	Status   string
	DateFrom string
	DateTo   string
}

func parseQueryParams(c *echo.Context) queryParams {
	values := c.Request().URL.Query()

	filter := queryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("campaign_id"),
		Type:     values.Get("type"),
		Label:    values.Get("label"),
		Status:   values.Get("status"),
		DateFrom: values.Get("date_from"),
		DateTo:   values.Get("date_to"),
	}

	return filter
}

func parseFilter(qp queryParams) (campaignbus.CampaignQueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter campaignbus.CampaignQueryFilter

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

	if qp.Type != "" {
		cmpType, err := campaignbus.ParseCampaignType(qp.Type)
		if err != nil {
			fieldErrors.Add("type", err)
		} else {
			filter.Type = &cmpType
		}
	}

	if qp.Status != "" {
		s, err := campaignbus.ParseCampaignStatus(qp.Status)
		if err != nil {
			fieldErrors.Add("status", err)
		} else {
			filter.Status = &s
		}
	}

	if qp.DateFrom != "" {
		t, err := time.Parse(time.RFC3339, qp.DateFrom)
		if err != nil {
			fieldErrors.Add("date_from", err)
		} else {
			filter.DateFrom = &t
		}
	}

	if qp.DateTo != "" {
		t, err := time.Parse(time.RFC3339, qp.DateTo)
		if err != nil {
			fieldErrors.Add("date_to", err)
		} else {
			filter.DateTo = &t
		}
	}

	if len(fieldErrors) > 0 {
		return campaignbus.CampaignQueryFilter{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return filter, nil
}
