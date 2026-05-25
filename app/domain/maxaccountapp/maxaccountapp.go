package maxaccountapp

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/query"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

type app struct {
	maxAccountBus *maxaccountbus.Business
}

func newApp(maxAccountBus *maxaccountbus.Business) *app {
	return &app{maxAccountBus: maxAccountBus}
}

func (a *app) beginLogin(c *echo.Context) error {
	var req BeginLoginRequest
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	newLogin, err := toBusBeginLogin(req)
	if err != nil {
		return err
	}

	attempt, err := a.maxAccountBus.BeginLogin(c.Request().Context(), newLogin)
	if err != nil {
		return mapBusErr(err, "beginlogin")
	}

	return c.JSON(http.StatusCreated, toAppLoginAttempt(attempt))
}

func (a *app) confirmLogin(c *echo.Context) error {
	var req ConfirmLoginRequest
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
	}

	confirm, err := toBusConfirmLogin(req, attemptID)
	if err != nil {
		return err
	}

	result, err := a.maxAccountBus.ConfirmLogin(c.Request().Context(), confirm)
	if err != nil {
		return mapBusErr(err, "confirmlogin")
	}

	return c.JSON(http.StatusOK, toAppLoginResult(result))
}

func (a *app) confirmPassword(c *echo.Context) error {
	var req ConfirmPasswordRequest
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
	}

	confirm, err := toBusConfirmPassword(req, attemptID)
	if err != nil {
		return err
	}

	result, err := a.maxAccountBus.ConfirmPassword(c.Request().Context(), confirm)
	if err != nil {
		return mapBusErr(err, "confirmpassword")
	}

	return c.JSON(http.StatusOK, toAppLoginResult(result))
}

func (a *app) query(c *echo.Context) error {
	qp := parseQueryParams(c)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err, errs.InvalidArgument, "invalid page")
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err
	}

	accounts, err := a.maxAccountBus.Query(c.Request().Context(), filter, page)
	if err != nil {
		return mapBusErr(err, "query")
	}

	count, err := a.maxAccountBus.Count(c.Request().Context(), filter)
	if err != nil {
		return mapBusErr(err, "count")
	}

	return c.JSON(http.StatusOK, query.NewResult(toAppAccounts(accounts), count, page))
}

func (a *app) update(c *echo.Context) error {
	var req UpdateAccountRequest
	if err := c.Bind(&req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateAccount(req)
	if err != nil {
		return err
	}

	account, err := a.accountByParam(c)
	if err != nil {
		return err
	}

	updated, err := a.maxAccountBus.Update(c.Request().Context(), account, up)
	if err != nil {
		return mapBusErr(err, "update")
	}
	return c.JSON(http.StatusOK, toAppAccount(updated))
}

func (a *app) start(c *echo.Context) error {
	account, err := a.accountByParam(c)
	if err != nil {
		return err
	}

	updated, err := a.maxAccountBus.Start(c.Request().Context(), account)
	if err != nil {
		return mapBusErr(err, "start")
	}
	return c.JSON(http.StatusOK, toAppAccount(updated))
}

func (a *app) stop(c *echo.Context) error {
	account, err := a.accountByParam(c)
	if err != nil {
		return err
	}

	updated, err := a.maxAccountBus.Stop(c.Request().Context(), account)
	if err != nil {
		return mapBusErr(err, "stop")
	}
	return c.JSON(http.StatusOK, toAppAccount(updated))
}

func (a *app) delete(c *echo.Context) error {
	account, err := a.accountByParam(c)
	if err != nil {
		return err
	}

	if err := a.maxAccountBus.Delete(c.Request().Context(), account); err != nil {
		return mapBusErr(err, "delete")
	}
	return c.NoContent(http.StatusNoContent)
}

func (a *app) accountByParam(c *echo.Context) (maxaccountbus.Account, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return maxaccountbus.Account{}, errs.New(errs.InvalidArgument, errs.ErrInvalidID)
	}

	account, err := a.maxAccountBus.QueryByID(c.Request().Context(), id)
	if err != nil {
		return maxaccountbus.Account{}, mapBusErr(err, "querybyid")
	}
	return account, nil
}
