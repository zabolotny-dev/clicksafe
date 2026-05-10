package visitapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type Config struct {
	Log      *logger.Logger
	VisitBus *visitbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.Log, cfg.VisitBus)

	router.GET("/:token", api.handleVisit)
}
