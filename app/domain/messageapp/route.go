package messageapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	MessageBus *messagebus.Business
	SessionBus *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.MessageBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadMessage := mid.LoadMessage(cfg.MessageBus)

	router.GET("/message", api.query, authen)
	router.POST("/message", api.create, authen, csrf)
	router.GET("/message/:id", api.queryByID, authen, loadMessage)
	router.PUT("/message/:id", api.update, authen, csrf, loadMessage)
	router.DELETE("/message/:id", api.deleteByID, authen, csrf, loadMessage)
}
