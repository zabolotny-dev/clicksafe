package attachmentapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
)

type Config struct {
	AttachmentBus *attachmentbus.Business
	RenderBus     *renderbus.Business
	SessionBus    *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.AttachmentBus, cfg.RenderBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadAttachment := mid.LoadAttachment(cfg.AttachmentBus)
	authorizeAttachment := mid.AuthorizeAttachment(cfg.AttachmentBus, cfg.SessionBus)

	router.POST("/attachment", api.save, authen, csrf)
	router.GET("/attachment/:id", api.download, authorizeAttachment)
	router.GET("/attachment", api.query, authen)
	router.PUT("/attachment/:id", api.update, authen, csrf, loadAttachment)
	router.PUT("/attachment/:id/content", api.updateContent, authen, csrf, loadAttachment)
	router.DELETE("/attachment/:id", api.deleteByID, authen, csrf, loadAttachment)
	router.GET("/attachment/:id/render/:target_id", api.render, authen, loadAttachment)
}
