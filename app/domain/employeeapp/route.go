package employeeapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
)

type Config struct {
	EmployeeBus *employeebus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.EmployeeBus)

	loadEmployee := mid.LoadEmployee(cfg.EmployeeBus)

	router.POST("/employee", api.create)
	router.GET("/employee/:id", api.queryByID, loadEmployee)
	router.GET("/employee", api.query)
	router.PUT("/employee/:id", api.update, loadEmployee)
	router.DELETE("/employee/:id", api.deleteByID, loadEmployee)
}
