package departmentapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	departmentBus *departmentbus.Business
}

func newApp(d *departmentbus.Business) *app {
	return &app{departmentBus: d}
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

	orderby, err := order.Parse(orderByFields, qp.OrderBy, departmentbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	deps, err := a.departmentBus.Query(c.Request().Context(), filter, orderby, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.departmentBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppDepartments(deps), count, page))
}

func (a *app) queryByID(c *echo.Context) error {
	dep, err := mid.GetDepartment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "querybyid: %s", err)
	}

	return c.JSON(http.StatusOK, toAppDepartment(dep))
}

func (a *app) create(c *echo.Context) error {
	var req NewDepartment
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	new, err := toBusNewDepartment(req)
	if err != nil {
		return err
	}

	if err := a.departmentBus.Save(c.Request().Context(), new); err != nil {
		return mapBusErr(err, "create")
	}

	return c.NoContent(http.StatusOK)
}

func (a *app) update(c *echo.Context) error {
	var req UpdateDepartment
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateDepartment(req)
	if err != nil {
		return err
	}

	dep, err := mid.GetDepartment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	if err = a.departmentBus.Update(c.Request().Context(), dep, up); err != nil {
		return mapBusErr(err, "update")
	}

	return c.NoContent(http.StatusOK)
}

func (a *app) deleteByID(c *echo.Context) error {
	dep, err := mid.GetDepartment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	err = a.departmentBus.Delete(c.Request().Context(), dep)
	if err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusOK)
}
