package vtargetapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
)

type Config struct {
	VTargetBus *vtargetbus.Business
	SessionBus *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.VTargetBus)

	authen := mid.Authenticate(cfg.SessionBus)

	router.GET("/vtarget", api.query, authen)
}
