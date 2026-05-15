package landingapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	LandingBus *landingbus.Business
	SessionBus *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.LandingBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadLanding := mid.LoadLanding(cfg.LandingBus)

	router.GET("/landing", api.query, authen)
	router.POST("/landing", api.create, authen, csrf)
	router.GET("/landing/:id", api.queryByID, authen, loadLanding)
	router.PUT("/landing/:id", api.update, authen, csrf, loadLanding)
	router.DELETE("/landing/:id", api.deleteByID, authen, csrf, loadLanding)
}
