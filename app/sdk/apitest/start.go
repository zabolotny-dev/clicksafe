package apitest

import (
	"io"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/api/service/build"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
)

// New initialises the full application stack backed by a real test database
// and returns a Test ready to run table-driven HTTP tests.
func New(t *testing.T, testName string) *Test {
	t.Helper()

	db := dbtest.New(t, testName)

	log := logger.New(io.Discard, logger.LevelError, "apitest")

	e := echo.New()

	build.Add(e, build.Config{
		Log:             log,
		OrganizationBus: db.BusDomain.Organization,
		DepartmentBus:   db.BusDomain.Department,
		EmployeeBus:     db.BusDomain.Employee,
		LandingBus:      db.BusDomain.Landing,
		MaxAccountBus:   db.BusDomain.MaxAccount,
		MessageBus:      db.BusDomain.Message,
		CampaignBus:     db.BusDomain.Campaign,
		TargetBus:       db.BusDomain.Target,
		VTargetBus:      db.BusDomain.VTarget,
		VisitBus:        db.BusDomain.Visit,
		AttachmentBus:   db.BusDomain.Attachment,
		RenderBus:       db.BusDomain.Render,
		VoiceAdapter:    nil, // no voice server in tests
		SessionBus:      db.BusDomain.Session,
		AuthBus:         db.BusDomain.Auth,
		LoginRateLimit: build.LoginRateLimitConfig{
			RequestsPerMinute: 1000,
			Burst:             1000,
			ExpiresIn:         time.Minute,
		},
	})

	return &Test{
		DB:  db,
		Mux: e,
	}
}
