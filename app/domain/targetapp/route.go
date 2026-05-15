package targetapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	TargetBus  *campaignbus.TargetBusiness
	SessionBus *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.TargetBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadTarget := mid.LoadTarget(cfg.TargetBus)

	router.POST("/target", api.create, authen, csrf)
	router.DELETE("/target/:id", api.deleteByID, authen, csrf, loadTarget)
	router.DELETE("/target/campaign/:campaign_id", api.deleteByCampaignID, authen, csrf)
	router.PUT("/target/:id/schedule", api.updateSchedule, authen, csrf, loadTarget)
	router.PUT("/target/campaign/:campaign_id/distribute", api.autoDistribute, authen, csrf)
}
