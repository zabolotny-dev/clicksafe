package eventapp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
)

type app struct {
	eventBus *eventbus.Business
}

func newApp(bus *eventbus.Business) *app {
	return &app{
		eventBus: bus,
	}
}

func (a *app) publish(c *echo.Context) error {
	var event Event
	if err := c.Bind(&event); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ev, err := toBusEvent(event)
	if err != nil {
		return err
	}

	if err := a.eventBus.Publish(c.Request().Context(), ev); err != nil {
		return mapBusErr(err, "publish")
	}

	return c.NoContent(http.StatusOK)
}
