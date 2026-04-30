package organizationapp

import (
	"errors"
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
	var org Organization
	if err := c.Bind(&org); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newOrg, err := toBusNewOrganization(org)
	if err != nil {
		return err
	}

	if err := a.organizationBus.Save(c.Request().Context(), newOrg); err != nil {
		return errs.Errorf(errs.InternalOnlyLog, "save: %s", err)
	}

	return nil
}

func (a *app) get(c *echo.Context) error {
	org, err := a.organizationBus.Get(c.Request().Context())
	if err != nil {
		if errors.Is(err, organizationbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Errorf(errs.InternalOnlyLog, "get: %s", err)
	}

	return c.JSON(http.StatusOK, org)
}

func (a *app) saveLogo(c *echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "file: %s", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "open: %s", err)
	}
	defer file.Close()

	ext := path.Ext(fileHeader.Filename)
	if ext == "" {
		return errs.Errorf(errs.InvalidArgument, "invalid extension")
	}

	logoURL, err := a.organizationBus.SaveLogo(c.Request().Context(), file, ext)
	if err != nil {
		if errors.Is(err, organizationbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Errorf(errs.InternalOnlyLog, "savelogo: %s", err)
	}

	return c.JSON(http.StatusOK, Logo{URL: logoURL})
}
