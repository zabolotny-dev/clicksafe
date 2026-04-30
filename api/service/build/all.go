package build

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/domain/eventapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/organizationapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type Config struct {
	Log             *logger.Logger
	EventBus        *eventbus.Business
	OrganizationBus *organizationbus.Business
}

func Add(e *echo.Echo, cfg Config) {
	e.HTTPErrorHandler = errs.NewEchoHandler(cfg.Log)

	eventapp.Routes(e, eventapp.Config{
		EventBus: cfg.EventBus,
	})

	organizationapp.Routes(e, organizationapp.Config{
		OrganizationBus: cfg.OrganizationBus,
	})
}
