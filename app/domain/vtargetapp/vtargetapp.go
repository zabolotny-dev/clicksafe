package vtargetapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	vtargetBus *vtargetbus.Business
}

func newApp(vtargetBus *vtargetbus.Business) *app {
	return &app{vtargetBus: vtargetBus}
}

func (a *app) query(c *echo.Context) error {
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, vtargetbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
	}

	targets, err := a.vtargetBus.Query(c.Request().Context(), filter, orderBy, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.vtargetBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppTargets(targets), count, page))
}
