package employeeapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	employeeBus *employeebus.Business
}

func newApp(d *employeebus.Business) *app {
	return &app{employeeBus: d}
}

func (a *app) create(c *echo.Context) error {
	var req NewEmployee
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	new, err := toBusNewEmployee(req)
	if err != nil {
		return err
	}

	res, err := a.employeeBus.Save(c.Request().Context(), new)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppEmployee(res))
}

func (a *app) queryByID(c *echo.Context) error {
	employee, err := mid.GetEmployee(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "querybyid: %s", err)
	}

	return c.JSON(http.StatusOK, toAppEmployee(employee))
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

	orderby, err := order.Parse(orderByFields, qp.OrderBy, employeebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	emps, err := a.employeeBus.Query(c.Request().Context(), filter, orderby, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.employeeBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppEmployees(emps), count, page))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateEmployee
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateEmployee(req)
	if err != nil {
		return err
	}

	emp, err := mid.GetEmployee(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	err = a.employeeBus.Update(c.Request().Context(), emp, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.NoContent(http.StatusOK)
}

func (a *app) deleteByID(c *echo.Context) error {
	employee, err := mid.GetEmployee(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "delete: %s", err)
	}

	err = a.employeeBus.Delete(c.Request().Context(), employee.ID)
	if err != nil {
		return mapBusErr(err, "delete")
	}

	return c.NoContent(http.StatusOK)
}
