package messageapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

type Config struct {
	MessageBus *messagebus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.MessageBus)

	loadMessage := mid.LoadMessage(cfg.MessageBus)

	router.GET("/message", api.query)
	router.POST("/message", api.create)
	router.GET("/message/:id", api.queryByID, loadMessage)
	router.PUT("/message/:id", api.update, loadMessage)
	router.DELETE("/message/:id", api.deleteByID, loadMessage)
}
