package mid

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
)

var ErrInvalidID = errors.New("ID is not in its proper form")

func LoadDepartment(departmentBus *departmentbus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, ErrInvalidID)
				}

				dep, err := departmentBus.QueryByID(c.Request().Context(), productID)
				if err != nil {
					switch {
					case errors.Is(err, departmentbus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "getbyid: departmentID[%s]: %s", productID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setDepartment(c.Request().Context(), dep),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}
