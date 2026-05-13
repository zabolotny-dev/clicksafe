package build

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/domain/attachmentapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/campaignapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/departmentapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/employeeapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/landingapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/messageapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/organizationapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/targetapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/visitapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/vtargetapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

type Config struct {
	Log             *logger.Logger
	OrganizationBus *organizationbus.Business
	DepartmentBus   *departmentbus.Business
	EmployeeBus     *employeebus.Business
	LandingBus      *landingbus.Business
	MessageBus      *messagebus.Business
	CampaignBus     *campaignbus.CampaignBusiness
	TargetBus       *campaignbus.TargetBusiness
	VTargetBus      *vtargetbus.Business
	VisitBus        *visitbus.Business
	AttachmentBus   *attachmentbus.Business
	RenderBus       *renderbus.Business
}

func Add(e *echo.Echo, cfg Config) {
	e.HTTPErrorHandler = errs.NewEchoHandler(cfg.Log)
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	organizationapp.Routes(e, organizationapp.Config{
		OrganizationBus: cfg.OrganizationBus,
	})

	departmentapp.Routes(e, departmentapp.Config{
		DepartmentBus: cfg.DepartmentBus,
	})

	employeeapp.Routes(e, employeeapp.Config{
		EmployeeBus: cfg.EmployeeBus,
	})

	messageapp.Routes(e, messageapp.Config{
		MessageBus: cfg.MessageBus,
	})

	landingapp.Routes(e, landingapp.Config{
		LandingBus: cfg.LandingBus,
	})

	campaignapp.Routes(e, campaignapp.Config{
		CampaignBus: cfg.CampaignBus,
	})

	targetapp.Routes(e, targetapp.Config{
		TargetBus: cfg.TargetBus,
	})

	vtargetapp.Routes(e, vtargetapp.Config{
		VTargetBus: cfg.VTargetBus,
	})

	attachmentapp.Routes(e, attachmentapp.Config{
		AttachmentBus: cfg.AttachmentBus,
		RenderBus:     cfg.RenderBus,
	})

	visitapp.Routes(e, visitapp.Config{
		Log:      cfg.Log,
		VisitBus: cfg.VisitBus,
	})
}
