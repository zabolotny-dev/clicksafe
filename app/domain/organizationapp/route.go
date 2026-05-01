package organizationapp

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

type Config struct {
	OrganizationBus *organizationbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.OrganizationBus)

	logoSizeLimit := middleware.BodyLimit(15 * 1024 * 1024)

	router.POST("/organization", api.create)
	router.GET("/organization", api.get)
	router.PUT("/organization/logo", api.saveLogo, logoSizeLimit, mid.AllowImagesOnly)
}
