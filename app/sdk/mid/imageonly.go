package mid

import (
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
)

func AllowImagesOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return errs.Errorf(errs.InvalidArgument, "file is required: %s", err)
		}
		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return errs.Errorf(errs.InvalidArgument, "only images are allowed, got: %s", contentType)
		}
		return next(c)
	}
}
