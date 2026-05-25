package maxaccountapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	MaxAccountBus *maxaccountbus.Business
	SessionBus    *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.MaxAccountBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()

	router.GET("/max-account", api.query, authen)
	router.POST("/max-account/login", api.beginLogin, authen, csrf)
	router.POST("/max-account/login/:attempt_id/confirm", api.confirmLogin, authen, csrf)
	router.POST("/max-account/login/:attempt_id/password", api.confirmPassword, authen, csrf)
	router.PUT("/max-account/:id", api.update, authen, csrf)
	router.POST("/max-account/:id/connect", api.start, authen, csrf)
	router.POST("/max-account/:id/disconnect", api.stop, authen, csrf)
	router.DELETE("/max-account/:id", api.delete, authen, csrf)
}
