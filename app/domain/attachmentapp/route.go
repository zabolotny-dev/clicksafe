package attachmentapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
)

type Config struct {
	AttachmentBus *attachmentbus.Business
	RenderBus     *renderbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.AttachmentBus, cfg.RenderBus)

	loadAttachment := mid.LoadAttachment(cfg.AttachmentBus)

	router.POST("/attachment", api.save)
	router.GET("/attachment/:id", api.download, loadAttachment)
	router.GET("/attachment", api.query)
	router.PUT("/attachment/:id", api.update, loadAttachment)
	router.DELETE("/attachment/:id", api.deleteByID, loadAttachment)
	router.GET("/attachment/:id/render/:target_id", api.render, loadAttachment)
}
