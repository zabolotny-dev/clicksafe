package departmentapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type Config struct {
	DepartmentBus *departmentbus.Business
	SessionBus    *sessionbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.DepartmentBus)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()
	loadDepartment := mid.LoadDepartment(cfg.DepartmentBus)

	router.GET("/department", api.query, authen)
	router.GET("/department/:id", api.queryByID, authen, loadDepartment)
	router.POST("/department", api.create, authen, csrf)
	router.PUT("/department/:id", api.update, authen, csrf, loadDepartment)
	router.DELETE("/department/:id", api.deleteByID, authen, csrf, loadDepartment)
}
