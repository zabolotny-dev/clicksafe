package mid

import (
	"crypto/subtle"
	"errors"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
)

func CSRF() echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			session, err := GetSession(c.Request().Context())
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			csrf := c.Request().Header.Get("X-CSRF-Token")
			if csrf == "" {
				return errs.New(errs.PermissionDenied, errors.New("missing csrf token"))
			}

			if subtle.ConstantTimeCompare([]byte(csrf), []byte(session.CSRFToken)) != 1 {
				return errs.New(errs.PermissionDenied, errors.New("invalid csrf token"))
			}

			return next(c)
		}

		return h
	}

	return m
}
