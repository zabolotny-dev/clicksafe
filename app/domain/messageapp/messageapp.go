package messageapp

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	messageBus *messagebus.Business
}

func newApp(messageBus *messagebus.Business) *app {
	return &app{messageBus: messageBus}
}

func (a *app) create(c *echo.Context) error {
	var req NewMessage
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newMsg, err := toBusNewMessage(req)
	if err != nil {
		return err
	}

	msg, err := a.messageBus.Save(c.Request().Context(), newMsg)
	if err != nil {
		return mapBusErr(err, "create")
	}

	return c.JSON(http.StatusCreated, toAppMessage(msg))
}

func (a *app) query(c *echo.Context) error {
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, messagebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
	}

	messages, err := a.messageBus.Query(c.Request().Context(), filter, orderBy, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.messageBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppMessages(messages), count, page))
}

func (a *app) queryByID(c *echo.Context) error {
	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "querybyid: %s", err)
	}

	return c.JSON(http.StatusOK, toAppMessage(msg))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateMessage
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateMessage(req)
	if err != nil {
		return err
	}

	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	updated, err := a.messageBus.Update(c.Request().Context(), msg, up)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppMessage(updated))
}

func (a *app) deleteByID(c *echo.Context) error {
	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	if err := a.messageBus.Delete(c.Request().Context(), msg); err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusOK)
}

func (a *app) saveContent(c *echo.Context) error {
	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "savecontent: %s", err)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "savecontent: file: %s", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "savecontent: open: %s", err)
	}
	defer file.Close()

	updated, err := a.messageBus.SaveContent(c.Request().Context(), msg, file)
	if err != nil {
		return mapBusErr(err, "savecontent")
	}

	return c.JSON(http.StatusOK, toAppMessage(updated))
}

func (a *app) readContent(c *echo.Context) error {
	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "readcontent: %s", err)
	}

	content, err := a.messageBus.ReadContent(c.Request().Context(), msg)
	if err != nil {
		return mapBusErr(err, "readcontent")
	}

	return c.Blob(http.StatusOK, "text/html; charset=utf-8", content)
}

func (a *app) render(c *echo.Context) error {
	var req RenderMessage
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	msg, err := mid.GetMessage(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "render: %s", err)
	}

	tID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return errs.NewFieldErrors("target_id", err, errs.InvalidArgument, "invalid target id")
	}

	content, err := a.messageBus.Render(c.Request().Context(), msg, tID)
	if err != nil {
		return mapBusErr(err, "render")
	}

	return c.JSON(http.StatusOK, RenderedMessage{Content: content})
}
