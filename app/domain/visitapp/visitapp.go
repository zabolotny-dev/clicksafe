package visitapp

import (
	"net/http"
	"net/netip"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type app struct {
	log      *logger.Logger
	visitBus *visitbus.Business
}

func newApp(log *logger.Logger, visitBus *visitbus.Business) *app {
	return &app{log: log, visitBus: visitBus}
}

func (a *app) serve(c *echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.HTML(http.StatusNotFound, "404 Not Found")
	}

	ipAddr, err := netip.ParseAddr(c.RealIP())
	if err != nil {
		a.log.Warn(c.Request().Context(), "invalid real ip",
			"ip", c.RealIP(),
			"remote_addr", c.Request().RemoteAddr,
			"x_forwarded_for", c.Request().Header.Get("X-Forwarded-For"),
			"err", err,
		)
	}
	data := visitbus.TargetData{
		Token:     token,
		IpAddress: ipAddr,
		UserAgent: c.Request().UserAgent(),
		Referer:   c.Request().Referer(),
	}

	html, err := a.visitBus.Serve(c.Request().Context(), data)
	if err != nil {
		a.log.Error(c.Request().Context(), "failed to serve landing", "err", err, "ip",
			data.IpAddress, "user_agent", data.UserAgent, "referer", data.Referer)
		return c.HTML(http.StatusNotFound, "404 Not Found")
	}

	return c.HTML(http.StatusOK, html)
}
