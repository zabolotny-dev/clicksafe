package campaignapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type Config struct {
	CampaignBus *campaignbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.CampaignBus)

	loadCampaign := mid.LoadCampaign(cfg.CampaignBus)

	router.POST("/campaign", api.create)
	router.GET("/campaign/:id", api.queryByID, loadCampaign)
	router.GET("/campaign", api.query)
	router.PUT("/campaign/:id", api.update, loadCampaign)
	router.DELETE("/campaign/:id", api.deleteByID, loadCampaign)
}
