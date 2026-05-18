package employeeapp

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/csv"
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
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderby, err := order.Parse(orderByFields, qp.OrderBy, employeebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
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

	updated, err := a.employeeBus.Update(c.Request().Context(), emp, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppEmployee(updated))
}

func (a *app) deleteByID(c *echo.Context) error {
	employee, err := mid.GetEmployee(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "delete: %s", err)
	}

	err = a.employeeBus.Delete(c.Request().Context(), employee)
	if err != nil {
		return mapBusErr(err, "delete")
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
		RequiredHeaders: []string{"first_name", "last_name", "email"},
		TrimValues:      true,
	})
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "importcsv: %s", err)
	}

	var fieldErrors errs.FieldErrors
	var newEmps []employeebus.NewEmployee
	rowsByEmail := make(map[string]int)
	rowsByPhone := make(map[string]int)

	for reader.Next() {
		row := reader.Row()

		newEmp, errs := csvToBusNewEmployee(row)
		if len(errs) > 0 {
			for _, rowErr := range errs {
				fieldErrors.AddValue("csv_row", rowErr)
			}
			continue
		}

		email := newEmp.Email.String()
		phone := newEmp.Phone.String()

		duplicated := false
		if firstRow, exists := rowsByEmail[email]; exists {
			fieldErrors.AddValue("csv_row", csvRowErrorValue{
				Row:   row.Number(),
				Field: "email",
				Err:   fmt.Sprintf("duplicate email, first seen on row %d", firstRow),
			})
			duplicated = true
		}

		if phone != "" {
			if firstRow, exists := rowsByPhone[phone]; exists {
				fieldErrors.AddValue("csv_row", csvRowErrorValue{
					Row:   row.Number(),
					Field: "phone",
					Err:   fmt.Sprintf("duplicate phone, first seen on row %d", firstRow),
				})
				duplicated = true
			}
		}

		if duplicated {
			continue
		}

		rowsByEmail[email] = row.Number()
		if phone != "" {
			rowsByPhone[phone] = row.Number()
		}

		newEmps = append(newEmps, newEmp)
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

	if err := a.employeeBus.SaveMany(c.Request().Context(), newEmps); err != nil {
		return mapBusErr(err, "importcsv")
	}

	return c.NoContent(http.StatusOK)
}
