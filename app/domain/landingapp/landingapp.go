package landingapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	landingBus *landingbus.Business
}

func newApp(landingBus *landingbus.Business) *app {
	return &app{landingBus: landingBus}
}

func (a *app) create(c *echo.Context) error {
	var req NewLanding
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newLanding, err := toBusNewLanding(req)
	if err != nil {
		return err
	}

	landing, err := a.landingBus.Save(c.Request().Context(), newLanding)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppLanding(landing))
}

func (a *app) query(c *echo.Context) error {
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, landingbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
	}

	landings, err := a.landingBus.Query(c.Request().Context(), filter, orderBy, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.landingBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppLandings(landings), count, page))
}

func (a *app) queryByID(c *echo.Context) error {
	landing, err := mid.GetLanding(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "querybyid: %s", err)
	}

	return c.JSON(http.StatusOK, toAppLanding(landing))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateLanding
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateLanding(req)
	if err != nil {
		return err
	}

	landing, err := mid.GetLanding(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	updated, err := a.landingBus.Update(c.Request().Context(), landing, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppLanding(updated))
}

func (a *app) deleteByID(c *echo.Context) error {
	landing, err := mid.GetLanding(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	if err := a.landingBus.Delete(c.Request().Context(), landing); err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusOK)
}
