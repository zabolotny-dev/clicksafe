package departmentapp

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
)

type Config struct {
	DepartmentBus *departmentbus.Business
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.DepartmentBus)

	loadDepartment := mid.LoadDepartment(cfg.DepartmentBus)

	router.GET("/department", api.query)
	router.GET("/department/:id", api.queryByID, loadDepartment)
	router.POST("/department", api.create)
	router.PUT("/department/:id", api.update, loadDepartment)
	router.DELETE("/department/:id", api.deleteByID, loadDepartment)
}
