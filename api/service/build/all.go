package build

import (
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/domain/attachmentapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/authapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/campaignapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/departmentapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/employeeapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/landingapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/maxaccountapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/messageapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/organizationapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/targetapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/visitapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/voiceapp"
	"github.com/zabolotny-dev/clicksafe/app/domain/vtargetapp"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/voiceadapter"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
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
	MaxAccountBus   *maxaccountbus.Business
	MessageBus      *messagebus.Business
	CampaignBus     *campaignbus.CampaignBusiness
	TargetBus       *campaignbus.TargetBusiness
	VTargetBus      *vtargetbus.Business
	VisitBus        *visitbus.Business
	AttachmentBus   *attachmentbus.Business
	RenderBus       *renderbus.Business
	VoiceAdapter    *voiceadapter.Client
	VoiceRuntime    voiceapp.RuntimeConfig
	SessionBus      *sessionbus.Business
	AuthBus         *authbus.Business
	LoginRateLimit  LoginRateLimitConfig
}

type LoginRateLimitConfig = mid.LoginRateLimitConfig
type VoiceRuntimeConfig = voiceapp.RuntimeConfig

func Add(e *echo.Echo, cfg Config) {
	e.HTTPErrorHandler = errs.NewEchoHandler(cfg.Log)
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Use(mid.NoCacheHeaders())

	organizationapp.Routes(e, organizationapp.Config{
		OrganizationBus: cfg.OrganizationBus,
		SessionBus:      cfg.SessionBus,
	})

	departmentapp.Routes(e, departmentapp.Config{
		DepartmentBus: cfg.DepartmentBus,
		SessionBus:    cfg.SessionBus,
	})

	employeeapp.Routes(e, employeeapp.Config{
		EmployeeBus: cfg.EmployeeBus,
		SessionBus:  cfg.SessionBus,
	})

	messageapp.Routes(e, messageapp.Config{
		MessageBus: cfg.MessageBus,
		SessionBus: cfg.SessionBus,
	})

	landingapp.Routes(e, landingapp.Config{
		LandingBus: cfg.LandingBus,
		SessionBus: cfg.SessionBus,
	})

	maxaccountapp.Routes(e, maxaccountapp.Config{
		MaxAccountBus: cfg.MaxAccountBus,
		SessionBus:    cfg.SessionBus,
	})

	campaignapp.Routes(e, campaignapp.Config{
		CampaignBus: cfg.CampaignBus,
		SessionBus:  cfg.SessionBus,
	})

	targetapp.Routes(e, targetapp.Config{
		TargetBus:  cfg.TargetBus,
		SessionBus: cfg.SessionBus,
	})

	vtargetapp.Routes(e, vtargetapp.Config{
		VTargetBus: cfg.VTargetBus,
		SessionBus: cfg.SessionBus,
	})

	attachmentapp.Routes(e, attachmentapp.Config{
		AttachmentBus: cfg.AttachmentBus,
		RenderBus:     cfg.RenderBus,
		SessionBus:    cfg.SessionBus,
	})

	voiceapp.Routes(e, voiceapp.Config{
		VoiceAdapterClient: cfg.VoiceAdapter,
		SessionBus:         cfg.SessionBus,
		Runtime:            cfg.VoiceRuntime,
	})

	authapp.Routes(e, authapp.Config{
		Log:            cfg.Log,
		AuthBus:        cfg.AuthBus,
		SessionBus:     cfg.SessionBus,
		LoginRateLimit: cfg.LoginRateLimit,
	})

	visitapp.Routes(e, visitapp.Config{
		Log:      cfg.Log,
		VisitBus: cfg.VisitBus,
	})
}
