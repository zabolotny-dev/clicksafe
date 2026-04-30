package eventapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
)

type Config struct {
	EventBus *eventbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.EventBus)

	router.POST("/events", api.publish)
}
