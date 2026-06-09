package attachmentapp

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
)

type app struct {
	attachmentBus *attachmentbus.Business
	renderBus     *renderbus.Business
}

func newApp(d *attachmentbus.Business, r *renderbus.Business) *app {
	return &app{attachmentBus: d, renderBus: r}
}

func (a *app) save(c *echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "save: file: %s", err)
	}

	base := filepath.Base(fileHeader.Filename)
	ext := strings.ToLower(filepath.Ext(base))
	name := strings.TrimSuffix(base, filepath.Ext(base))

	var public bool
	publicRaw := c.FormValue("public")
	if publicRaw != "" {
		public, err = strconv.ParseBool(publicRaw)
		if err != nil {
			return errs.NewFieldErrors("public", err, errs.InvalidArgument, "invalid public")
		}
	}

	attach, err := toBusNewAttachment(name, public)
	if err != nil {
		return err
	}

	atchType, err := attachmentbus.Parse(ext)
	if err != nil {
		return errs.New(errs.InvalidArgument, fmt.Errorf("save: %w", err))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "save: open: %s", err)
	}
	defer file.Close()

	attach.Type = atchType
	attach.Content = file

	res, err := a.attachmentBus.Save(c.Request().Context(), attach)
	if err != nil {
		return mapBusErr(err, "save")
	}

	return c.JSON(http.StatusOK, toAppAttachment(res, c.Path()))
}

func (a *app) query(c *echo.Context) error {
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, attachmentbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err, errs.InvalidArgument, "invalid order")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
	}

	attachments, err := a.attachmentBus.Query(c.Request().Context(), filter, orderBy, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.attachmentBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppAttachments(attachments, c.Path()), count, page))
}

func (a *app) update(c *echo.Context) error {
	var up UpdateAttachment
	if err := c.Bind(&up); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	busUp, err := toBusUpdateAttachment(up)
	if err != nil {
		return err
	}

	attach, err := mid.GetAttachment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "update: %s", err)
	}

	res, err := a.attachmentBus.Update(c.Request().Context(), attach, busUp)
	if err != nil {
		return mapBusErr(err, "update")
	}

	return c.JSON(http.StatusOK, toAppAttachment(res, c.Path()))
}

func (a *app) updateContent(c *echo.Context) error {
	var up UpdateAttachmentContent
	if err := c.Bind(&up); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if up.Content == nil {
		return errs.NewFieldErrors("content", fmt.Errorf("content is required"), errs.InvalidArgument, "validation failed")
	}

	attach, err := mid.GetAttachment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "updatecontent: %s", err)
	}

	res, err := a.attachmentBus.UpdateContent(c.Request().Context(), attach, attachmentbus.UpdateAttachmentContent{
		Content: strings.NewReader(*up.Content),
	})
	if err != nil {
		return mapBusErr(err, "updatecontent")
	}

	return c.JSON(http.StatusOK, toAppAttachment(res, c.Path()))
}

func (a *app) deleteByID(c *echo.Context) error {
	atch, err := mid.GetAttachment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "deletebyid: %s", err)
	}

	if err := a.attachmentBus.Delete(c.Request().Context(), atch); err != nil {
		return mapBusErr(err, "deletebyid")
	}

	return c.NoContent(http.StatusNoContent)
}

func (a *app) download(c *echo.Context) error {
	atch, err := mid.GetAttachment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "download: %s", err)
	}

	content, err := a.attachmentBus.OpenContent(c.Request().Context(), atch)
	if err != nil {
		return mapBusErr(err, "download")
	}
	defer content.Close()

	filename := atch.Label.String() + atch.Type.String()
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	})

	c.Response().Header().Set(echo.HeaderContentDisposition, disposition)
	c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")

	return c.Stream(http.StatusOK, attachmentContentType(atch.Type), content)
}

func (a *app) render(c *echo.Context) error {
	targetIDStr := c.Param("target_id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		return errs.NewFieldErrors("target_id", err, errs.InvalidArgument, "invalid target_id")
	}

	atch, err := mid.GetAttachment(c.Request().Context())
	if err != nil {
		return errs.Errorf(errs.Internal, "render: %s", err)
	}

	content, err := a.renderBus.Render(c.Request().Context(), atch, targetID)
	if err != nil {
		return mapBusErr(err, "render")
	}

	filename := atch.Label.String() + "_rendered" + atch.Type.String()
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	})

	c.Response().Header().Set(echo.HeaderContentDisposition, disposition)
	c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")

	return c.Stream(http.StatusOK, attachmentContentType(atch.Type), bytes.NewReader(content))
}
