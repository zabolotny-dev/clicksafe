package visitapp

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type app struct {
	log      *logger.Logger
	visitBus *visitbus.Business
}

var trackingPixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

func newApp(log *logger.Logger, visitBus *visitbus.Business) *app {
	return &app{log: log, visitBus: visitBus}
}

func (a *app) handleVisit(c *echo.Context) error {
	token := c.Param("token")

	if cleanToken, ok := strings.CutSuffix(token, ".gif"); ok {
		return a.trackOpen(c, cleanToken)
	}

	return a.serveLanding(c, token)
}

func (a *app) serveLanding(c *echo.Context, token string) error {
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

func (a *app) trackOpen(c *echo.Context, token string) error {
	if token != "" {
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

		if err := a.visitBus.TrackOpen(c.Request().Context(), data); err != nil {
			a.log.Error(c.Request().Context(), "failed to track email open", "err", err, "ip",
				data.IpAddress, "user_agent", data.UserAgent, "referer", data.Referer)
		}
	}

	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")

	return c.Blob(http.StatusOK, "image/gif", trackingPixelGIF)
}
