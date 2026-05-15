package organizationapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	OrganizationBus *organizationbus.Business
	SessionBus      *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.OrganizationBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()

	router.POST("/organization", api.create, authen, csrf)
	router.GET("/organization", api.get, authen)
	router.PUT("/organization", api.update, authen, csrf)
}
