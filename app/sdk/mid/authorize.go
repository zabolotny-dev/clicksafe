package mid

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

func AuthorizeAttachment(attachmentBus *attachmentbus.Business, sessionBus *sessionbus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				attachmentID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				attachment, err := attachmentBus.QueryByID(c.Request().Context(), attachmentID)
				if err != nil {
					switch {
					case errors.Is(err, attachmentbus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: attachmentID[%s]: %s", attachmentID, err)
					}
				}

				reqCtx := setAttachment(c.Request().Context(), attachment)

				if !attachment.Public {
					cookie, err := c.Cookie("__Host-session")
					if err != nil {
						return errs.New(errs.Unauthenticated, err)
					}

					session, err := sessionBus.Authenticate(c.Request().Context(), cookie.Value)
					if err != nil {
						return errs.New(errs.Unauthenticated, err)
					}

					reqCtx = setSession(reqCtx, session)
				}

				c.SetRequest(c.Request().WithContext(reqCtx))
			}

			return next(c)
		}

		return h
	}

	return m
}
