package dbtest

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	admindbstore "github.com/zabolotny-dev/clicksafe/business/domain/adminbus/stores/admindb"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	attachmentdbstore "github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus/stores/attachmentdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	campaigndbstore "github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/campaigndb"
	targetdbstore "github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/targetdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	departmentdbstore "github.com/zabolotny-dev/clicksafe/business/domain/departmentbus/stores/departmentdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	employeedbstore "github.com/zabolotny-dev/clicksafe/business/domain/employeebus/stores/employeedb"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	eventdbstore "github.com/zabolotny-dev/clicksafe/business/domain/eventbus/stores/eventdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	landingdbstore "github.com/zabolotny-dev/clicksafe/business/domain/landingbus/stores/landingdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	maxaccountdbstore "github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus/stores/maxaccountdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	messagedbstore "github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	organizationdbstore "github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	sessiondbstore "github.com/zabolotny-dev/clicksafe/business/domain/sessionbus/stores/sessiondb"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	vtargetdbstore "github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus/stores/vtargetdb"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/password"
	"github.com/zabolotny-dev/clicksafe/foundation/token"
)

// BusDomain contains all business domain packages needed for integration testing.
type BusDomain struct {
	Admin        *adminbus.Business
	Session      *sessionbus.Business
	Auth         *authbus.Business
	Organization *organizationbus.Business
	Department   *departmentbus.Business
	Employee     *employeebus.Business
	Campaign     *campaignbus.CampaignBusiness
	Target       *campaignbus.TargetBusiness
	VTarget      *vtargetbus.Business
	Message      *messagebus.Business
	Landing      *landingbus.Business
	Attachment   *attachmentbus.Business
	MaxAccount   *maxaccountbus.Business
	Render       *renderbus.Business
	Visit        *visitbus.Business
	Resolver     *resolverbus.Business
	Event        *eventbus.Business
}

// testMaxAdapter is a no-op adapter used in tests so maxaccountbus can be wired
// without a real MAX gRPC server.
type testMaxAdapter struct{}

func (testMaxAdapter) BeginLogin(_ context.Context, _ string) (maxaccountbus.LoginAttempt, error) {
	return maxaccountbus.LoginAttempt{}, nil
}
func (testMaxAdapter) ConfirmLogin(_ context.Context, _, _, _ string) (maxaccountbus.LoginResult, error) {
	return maxaccountbus.LoginResult{}, nil
}
func (testMaxAdapter) ConfirmPassword(_ context.Context, _, _ string) (maxaccountbus.LoginResult, error) {
	return maxaccountbus.LoginResult{}, nil
}
func (testMaxAdapter) ListAccounts(_ context.Context) ([]maxaccountbus.Account, error) {
	return nil, nil
}
func (testMaxAdapter) StartAccount(_ context.Context, _ uuid.UUID) (maxaccountbus.Account, error) {
	return maxaccountbus.Account{}, nil
}
func (testMaxAdapter) StopAccount(_ context.Context, _ uuid.UUID) (maxaccountbus.Account, error) {
	return maxaccountbus.Account{}, nil
}
func (testMaxAdapter) DeleteAccount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func newBusDomains(pool *pgxpool.Pool, tmpDir string) BusDomain {
	hasher, err := password.New(password.Argon2idConfig{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		panic("dbtest: creating password hasher: " + err.Error())
	}

	tokenMgr, err := token.New(token.Config{
		TokenBytes: 32,
		HashKey:    "test-key-for-dbtest-32-bytes-xxx",
	})
	if err != nil {
		panic("dbtest: creating token manager: " + err.Error())
	}

	fs := filestore.New(tmpDir, "/attachment")

	adminStore := admindbstore.NewStore(pool)
	adminBus := adminbus.NewBusiness(hasher, adminStore)

	sessionStore := sessiondbstore.NewStore(pool)
	sessionBus := sessionbus.NewBusiness(sessionStore, tokenMgr, 8*time.Hour)

	authBus := authbus.NewBusiness(adminBus, sessionBus)

	eventStore := eventdbstore.NewStore(pool)
	eventBus := eventbus.NewBusinnes(eventStore)

	attachmentStore := attachmentdbstore.NewStore(pool)
	attachmentBus := attachmentbus.NewBusiness(fs, attachmentStore)

	organizationStore := organizationdbstore.NewStore(pool)
	organizationBus := organizationbus.NewBusiness(organizationStore, attachmentBus)

	departmentStore := departmentdbstore.NewStore(pool)
	departmentBus := departmentbus.NewBusiness(departmentStore)

	employeeStore := employeedbstore.NewStore(pool)
	employeeBus := employeebus.NewBusiness(employeeStore)

	vtargetStore := vtargetdbstore.NewStore(pool)
	vtargetBus := vtargetbus.NewBusiness(vtargetStore)

	targetStore := targetdbstore.NewStore(pool)
	campaignStore := campaigndbstore.NewStore(pool)
	targetBus := campaignbus.NewTargetBusiness(campaignStore, targetStore)

	messageStore := messagedbstore.NewStore(pool)
	messageBus := messagebus.NewBusiness(messageStore, attachmentBus)

	landingStore := landingdbstore.NewStore(pool)
	landingBus := landingbus.NewBusiness(landingStore, attachmentBus)

	resolverBus := resolverbus.NewBusiness(targetBus, employeeBus, departmentBus, organizationBus, attachmentBus)

	renderBus := renderbus.NewBusiness(fs, resolverBus)

	campaignBus := campaignbus.NewCampaignBusiness(
		campaignStore, targetStore,
		messageBus, landingBus, employeeBus,
		resolverBus, attachmentBus,
	)

	maxAccountStore := maxaccountdbstore.NewStore(pool)
	maxAccountBus := maxaccountbus.NewBusiness(maxAccountStore, testMaxAdapter{})

	visitBus := visitbus.NewBusiness(targetBus, campaignBus, landingBus, eventBus, attachmentBus, renderBus)

	return BusDomain{
		Admin:        adminBus,
		Session:      sessionBus,
		Auth:         authBus,
		Organization: organizationBus,
		Department:   departmentBus,
		Employee:     employeeBus,
		Campaign:     campaignBus,
		Target:       targetBus,
		VTarget:      vtargetBus,
		Message:      messageBus,
		Landing:      landingBus,
		Attachment:   attachmentBus,
		MaxAccount:   maxAccountBus,
		Render:       renderBus,
		Visit:        visitBus,
		Resolver:     resolverBus,
		Event:        eventBus,
	}
}
