package employeeapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	EmployeeBus *employeebus.Business
	SessionBus  *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.EmployeeBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadEmployee := mid.LoadEmployee(cfg.EmployeeBus)

	router.POST("/employee", api.create, authen, csrf)
	router.GET("/employee/:id", api.queryByID, authen, loadEmployee)
	router.GET("/employee", api.query, authen)
	router.PUT("/employee/:id", api.update, authen, csrf, loadEmployee)
	router.DELETE("/employee/:id", api.deleteByID, authen, csrf, loadEmployee)
}
