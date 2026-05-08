package visitapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/visitbus"
)

type app struct {
	visitBus *visitbus.Business
}

func newApp(visitBus *visitbus.Business) *app {
	return &app{visitBus: visitBus}
}

func (a *app) serve(c *echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.HTML(http.StatusNotFound, "404 Not Found")
	}

	html, err := a.visitBus.Serve(c.Request().Context(), token)
	if err != nil {
		return c.HTML(http.StatusNotFound, "404 Not Found")
	}

	return c.HTML(http.StatusOK, html)
}
