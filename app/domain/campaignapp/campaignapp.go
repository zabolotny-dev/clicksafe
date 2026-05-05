package campaignapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	campaignBus *campaignbus.Business
}

func newApp(d *campaignbus.Business) *app {
	return &app{campaignBus: d}
}

func (a *app) create(c *echo.Context) error {
	var req NewCampaign
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newCampaign, err := toBusNewCampaign(req)
	if err != nil {
		return err
	}

	cmp, err := a.campaignBus.Save(c.Request().Context(), newCampaign)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppCampaign(cmp))
}

func (a *app) query(c *echo.Context) error {
	qp, err := parseQueryParams(c)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, campaignbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	campaigns, err := a.campaignBus.Query(c.Request().Context(), filter, orderBy, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.campaignBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppCampaigns(campaigns), count, page))
}

func (a *app) queryByID(c *echo.Context) error {
	cmp, err := mid.GetCampaign(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "querybyid: %s", err)
	}

	return c.JSON(http.StatusOK, toAppCampaign(cmp))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateCampaign
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateCampaign(req)
	if err != nil {
		return err
	}

	cmp, err := mid.GetCampaign(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	updated, err := a.campaignBus.Update(c.Request().Context(), cmp, up)
	if err != nil {
		return mapBusErr(err, "update")
	}
	return c.JSON(http.StatusOK, toAppCampaign(updated))
}

func (a *app) deleteByID(c *echo.Context) error {
	cmp, err := mid.GetCampaign(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	if err := a.campaignBus.Delete(c.Request().Context(), cmp); err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusNoContent)
}
