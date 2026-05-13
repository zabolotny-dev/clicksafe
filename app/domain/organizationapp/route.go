package organizationapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

type Config struct {
	OrganizationBus *organizationbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.OrganizationBus)

	router.POST("/organization", api.create)
	router.GET("/organization", api.get)
	router.PUT("/organization", api.update)
}
