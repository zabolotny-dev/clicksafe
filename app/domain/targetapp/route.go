package targetapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type Config struct {
	TargetBus *campaignbus.TargetBusiness
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.TargetBus)

	loadTarget := mid.LoadTarget(cfg.TargetBus)

	router.POST("/target", api.create)
	router.DELETE("/target/:id", api.deleteByID, loadTarget)
	router.DELETE("/target/campaign/:campaign_id", api.deleteByCampaignID)
	router.PUT("/target/:id/schedule", api.updateSchedule, loadTarget)
	router.PUT("/target/campaign/:campaign_id/distribute", api.autoDistribute)
}
