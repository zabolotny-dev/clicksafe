package mid

import "github.com/labstack/echo/v5"

func NoCacheHeaders() echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			headers := c.Response().Header()
			headers.Set("Cache-Control", "no-store")
			headers.Set("Pragma", "no-cache")
			headers.Set("Expires", "0")

			return next(c)
		}

		return h
	}

	return m
}
