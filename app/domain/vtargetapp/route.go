package vtargetapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
)

type Config struct {
	VTargetBus *vtargetbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.VTargetBus)

	router.GET("/vtarget", api.query)
}
