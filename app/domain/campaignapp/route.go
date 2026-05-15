package campaignapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	CampaignBus *campaignbus.CampaignBusiness
	SessionBus  *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.CampaignBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadCampaign := mid.LoadCampaign(cfg.CampaignBus)

	router.POST("/campaign", api.create, authen, csrf)
	router.GET("/campaign/:id", api.queryByID, authen, loadCampaign)
	router.GET("/campaign", api.query, authen)
	router.PUT("/campaign/:id", api.update, authen, csrf, loadCampaign)
	router.PUT("/campaign/:id/start", api.start, authen, csrf, loadCampaign)
	router.PUT("/campaign/:id/pause", api.pause, authen, csrf, loadCampaign)
	router.PUT("/campaign/:id/cancel", api.cancel, authen, csrf, loadCampaign)
	router.DELETE("/campaign/:id", api.deleteByID, authen, csrf, loadCampaign)
}
