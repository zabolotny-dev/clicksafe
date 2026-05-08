package landingapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
)

type Config struct {
	LandingBus *landingbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.LandingBus)

	loadLanding := mid.LoadLanding(cfg.LandingBus)

	router.GET("/landing", api.query)
	router.POST("/landing", api.create)
	router.PUT("/landing/:id/content", api.saveContent, loadLanding)
	router.GET("/landing/:id/content", api.readContent, loadLanding)
	router.POST("/landing/:id/render", api.render, loadLanding)
	router.GET("/landing/:id", api.queryByID, loadLanding)
	router.PUT("/landing/:id", api.update, loadLanding)
	router.DELETE("/landing/:id", api.deleteByID, loadLanding)
}
