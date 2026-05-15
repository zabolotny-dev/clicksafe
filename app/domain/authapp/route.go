package authapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type Config struct {
	Log            *logger.Logger
	AuthBus        *authbus.Business
	SessionBus     *sessionbus.Business
	LoginRateLimit mid.LoginRateLimitConfig
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.Log, cfg.AuthBus)

	authen := mid.Authenticate(cfg.SessionBus)
	loginLimiter := mid.RateLimitLogin(cfg.LoginRateLimit)

	router.POST("/login", api.login, loginLimiter)
	router.POST("/logout", api.logout, authen)
	router.GET("/me", api.me, authen)
}
