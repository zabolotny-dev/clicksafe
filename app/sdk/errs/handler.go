package errs

import (
	"errors"
	"path"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

func NewEchoHandler(log *logger.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		var appErr *Error
		isAppErr := errors.As(err, &appErr)

		logArgs := []any{"err", err}
		if isAppErr {
			logArgs = append(logArgs,
				"source_err_file", path.Base(appErr.FileName),
				"source_err_func", path.Base(appErr.FuncName),
			)
		}

		log.Error(c.Request().Context(), "handled error during request", logArgs...)

		if isAppErr {
			if appErr.Code == InternalOnlyLog {
				appErr = Errorf(Internal, "Internal Server Error")
			}
			c.JSON(appErr.HTTPStatus(), appErr)
			return
		}

		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) {
			appErr = Errorf(FromHTTPStatus(sc.StatusCode()), "%s", err.Error())
		} else {
			appErr = Errorf(Internal, "Internal Server Error")
		}

		c.JSON(appErr.HTTPStatus(), appErr)
	}
}
