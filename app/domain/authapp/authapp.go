package authapp

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type app struct {
	authBus *authbus.Business
	log     *logger.Logger
}

func newApp(log *logger.Logger, ab *authbus.Business) *app {
	return &app{log: log, authBus: ab}
}

func (a *app) login(c *echo.Context) error {
	var req Login
	if err := c.Bind(&req); err != nil {
		return err
	}

	ld, err := toBusLoginData(req)
	if err != nil {
		return err
	}

	userAgent := c.Request().UserAgent()
	ipAddr, err := netip.ParseAddr(c.RealIP())
	if err != nil {
		a.log.Warn(c.Request().Context(), "invalid real ip",
			"ip", c.RealIP(),
			"remote_addr", c.Request().RemoteAddr,
			"x_forwarded_for", c.Request().Header.Get("X-Forwarded-For"),
			"err", err,
		)
	}
	ld.UserAgent = userAgent
	ld.IPAddress = ipAddr

	res, err := a.authBus.Login(c.Request().Context(), ld)
	if err != nil {
		return mapBusErr(err, "login")
	}

	c.SetCookie(&http.Cookie{
		Name:     "__Host-session",
		Value:    res.Session.Token,
		Path:     "/",
		Expires:  res.Session.ExpiresAt,
		MaxAge:   int(time.Until(res.Session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, LoginResult{
		FirstName: res.Admin.FirstName.String(),
		LastName:  res.Admin.LastName.String(),
		CSRFToken: res.Session.CSRFToken,
		ExpiresAt: res.Session.ExpiresAt,
	})
}

func (a *app) logout(c *echo.Context) error {
	session, err := mid.GetSession(c.Request().Context())
	if err != nil {
		return errs.New(errs.Internal, fmt.Errorf("logout: %w", err))
	}

	err = a.authBus.Logout(c.Request().Context(), session)
	if err != nil {
		return mapBusErr(err, "logout")
	}

	c.SetCookie(&http.Cookie{
		Name:     "__Host-session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.NoContent(http.StatusNoContent)
}

func (a *app) me(c *echo.Context) error {
	session, err := mid.GetSession(c.Request().Context())
	if err != nil {
		return errs.New(errs.Internal, fmt.Errorf("me: %w", err))
	}

	adminInfo, err := a.authBus.Me(c.Request().Context(), session)
	if err != nil {
		return mapBusErr(err, "me")
	}

	return c.JSON(http.StatusOK, Me{
		FirstName: adminInfo.FirstName.String(),
		LastName:  adminInfo.LastName.String(),
		CSRFToken: session.CSRFToken,
		ExpiresAt: session.ExpiresAt,
	})
}
