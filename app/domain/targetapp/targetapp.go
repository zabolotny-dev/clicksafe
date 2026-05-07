package targetapp

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type app struct {
	targetBus *campaignbus.TargetBusiness
}

func newApp(d *campaignbus.TargetBusiness) *app {
	return &app{targetBus: d}
}

func (a *app) create(c *echo.Context) error {
	var req NewTarget
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newTarget, err := toBusNewTarget(req)
	if err != nil {
		return err
	}

	target, err := a.targetBus.Save(c.Request().Context(), newTarget)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppTarget(target))
}

func (a *app) deleteByID(c *echo.Context) error {
	target, err := mid.GetTarget(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	if err := a.targetBus.Delete(c.Request().Context(), target); err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusNoContent)
}

func (a *app) deleteByCampaignID(c *echo.Context) error {
	campaignID, err := uuid.Parse(c.Param("campaign_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
	}

	if err := a.targetBus.DeleteByCampaignID(c.Request().Context(), campaignID); err != nil {
		return mapBusErr(err, "deletebycampaignid")
	}

	return c.NoContent(http.StatusNoContent)
}

func (a *app) updateSchedule(c *echo.Context) error {
	var req UpdateSchedule
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	scheduledAt, err := toBusUpdateSchedule(req)
	if err != nil {
		return err
	}

	target, err := mid.GetTarget(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "updateschedule: %s", err)
	}

	updated, err := a.targetBus.UpdateSchedule(c.Request().Context(), target, scheduledAt)
	if err != nil {
		return mapBusErr(err, "updateschedule")
	}

	return c.JSON(http.StatusOK, toAppTarget(updated))
}

func (a *app) autoDistribute(c *echo.Context) error {
	campaignID, err := uuid.Parse(c.Param("campaign_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
	}

	if err := a.targetBus.AutoDistribute(c.Request().Context(), campaignID); err != nil {
		return mapBusErr(err, "autodistribute")
	}

	return c.NoContent(http.StatusNoContent)
}
