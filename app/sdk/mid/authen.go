package mid

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

func Authenticate(sb *sessionbus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			cookie, err := c.Cookie("__Host-session")
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			session, err := sb.Authenticate(c.Request().Context(), cookie.Value)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			c.SetRequest(c.Request().WithContext(
				setSession(c.Request().Context(), session),
			))

			return next(c)
		}
		return h
	}
	return m
}
