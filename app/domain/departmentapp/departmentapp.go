package departmentapp

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/csv"
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
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderby, err := order.Parse(orderByFields, qp.OrderBy, departmentbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
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

	dep, err := a.departmentBus.Save(c.Request().Context(), new)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppDepartment(dep))
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

	updated, err := a.departmentBus.Update(c.Request().Context(), dep, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppDepartment(updated))
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

func (a *app) importCSV(c *echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "importcsv: %s", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "importcsv: %s", err)
	}
	defer file.Close()

	reader, err := csv.NewReader(file, csv.Config{
		RequiredHeaders: []string{"label"},
		TrimValues:      true,
	})
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "importcsv: %s", err)
	}

	var fieldErrors errs.FieldErrors
	departments := make([]departmentbus.NewDepartment, 0)
	rowsByLabel := make(map[string]int)

	for reader.Next() {
		row := reader.Row()

		new, rowErrors := csvToBusNewDepartment(row)
		if len(rowErrors) > 0 {
			for _, rowErr := range rowErrors {
				fieldErrors.AddValue("csv_row", rowErr)
			}
			continue
		}

		if firstRow, exists := rowsByLabel[new.Label.String()]; exists {
			fieldErrors.AddValue("csv_row", csvRowErrorValue{
				Row:   row.Number(),
				Field: "label",
				Err:   fmt.Sprintf("duplicate label, first seen on row %d", firstRow),
			})
			continue
		}

		rowsByLabel[new.Label.String()] = row.Number()
		departments = append(departments, new)
	}

	for _, csvErr := range reader.Err() {
		fieldErrors.AddValue("csv_row", csvRowErrorValue{
			Row: csvErr.Row,
			Err: csvErr.Err,
		})
	}

	if len(fieldErrors) > 0 {
		return fieldErrors.ToError(errs.InvalidArgument, "invalid csv")
	}

	if err := a.departmentBus.SaveMany(c.Request().Context(), departments); err != nil {
		return mapBusErr(err, "importcsv")
	}

	return c.NoContent(http.StatusOK)
}
