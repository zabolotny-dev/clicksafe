package organizationapp

import (
	"net/http"
	"path"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
)

type app struct {
	organizationBus *organizationbus.Business
}

func newApp(ob *organizationbus.Business) *app {
	return &app{organizationBus: ob}
}

func (a *app) create(c *echo.Context) error {
	var req NewOrganization
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newOrg, err := toBusNewOrganization(req)
	if err != nil {
		return err
	}

	org, err := a.organizationBus.Save(c.Request().Context(), newOrg)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppOrganization(org))
}

func (a *app) get(c *echo.Context) error {
	org, err := a.organizationBus.Get(c.Request().Context())
	if err != nil {
		return mapBusErr(err, "get")
	}

	return c.JSON(http.StatusOK, toAppOrganization(org))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateOrganization
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateOrganization(req)
	if err != nil {
		return err
	}

	org, err := a.organizationBus.Get(c.Request().Context())
	if err != nil {
		return mapBusErr(err, "update")
	}

	updated, err := a.organizationBus.Update(c.Request().Context(), org, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppOrganization(updated))
}

func (a *app) updateLogo(c *echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "updatelogo: file: %s", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "updatelogo: open: %s", err)
	}
	defer file.Close()

	ext := path.Ext(fileHeader.Filename)
	if ext == "" {
		return errs.Errorf(errs.InvalidArgument, "updatelogo: invalid extension")
	}

	logoPath, err := a.organizationBus.UpdateLogo(c.Request().Context(), file, ext)
	if err != nil {
		return mapBusErr(err, "updatelogo")
	}

	return c.JSON(http.StatusOK, Logo{Path: logoPath})
}

func (a *app) deleteLogo(c *echo.Context) error {
	if err := a.organizationBus.DeleteLogo(c.Request().Context()); err != nil {
		return mapBusErr(err, "deletelogo")
	}

	return c.NoContent(http.StatusNoContent)
}
