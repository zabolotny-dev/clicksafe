package visitapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/visitbus"
)

type Config struct {
	VisitBus *visitbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.VisitBus)

	router.GET("/:token", api.serve)
}
