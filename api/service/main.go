package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/api/service/build"
	"github.com/zabolotny-dev/clicksafe/api/service/workers"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus/stores/admindb"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus/stores/attachmentdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/campaigndb"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/targetdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus/stores/departmentdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus/stores/employeedb"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus/stores/eventdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus/stores/landingdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus/stores/maxaccountdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus/stores/sessiondb"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus/stores/vtargetdb"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/sdk/maxadapter"
	"github.com/zabolotny-dev/clicksafe/business/sdk/voiceadapter"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/deliverybus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus/stores/maxdeliverydb"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
	"github.com/zabolotny-dev/clicksafe/foundation/mail"
	"github.com/zabolotny-dev/clicksafe/foundation/password"
	"github.com/zabolotny-dev/clicksafe/foundation/token"
)

func main() {
	ctx := context.Background()

	log := logger.New(os.Stdout, logger.LevelInfo, "api")

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		App struct {
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
		Web struct {
			APIHost      string        `conf:"default:0.0.0.0:8080"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:vladick,mask"`
			Host         string `conf:"default:localhost:5432"`
			Name         string `conf:"default:clicksafe"`
			MaxOpenConns int    `conf:"default:25"`
		}
		Storage struct {
			AttachmentRootDir    string `conf:"default:./private/attachment"`
			AttachmentPathPrefix string `conf:"default:/attachment"`
		}
		Worker struct {
			Interval time.Duration `conf:"default:1m"`
		}
		SMTP struct {
			Host     string        `conf:"default:localhost"`
			Port     int           `conf:"default:1025"`
			Username string        `conf:"noprint"`
			Password string        `conf:"noprint"`
			Timeout  time.Duration `conf:"default:10s"`
			TLS      string        `conf:"default:none"`
			SSL      bool          `conf:"default:false"`
		}
		MaxAdapter struct {
			Addr  string `conf:"default:localhost:9090"`
			Token string `conf:"noprint"`
		}
		VoiceAdapter struct {
			Addr  string `conf:"default:localhost:9091"`
			Token string `conf:"noprint"`
		}
		Voice struct {
			CloneTimeout    time.Duration `conf:"default:10m,env:VOICE_CLONE_TIMEOUT"`
			ASRTimeout      time.Duration `conf:"default:5m,env:VOICE_ASR_TIMEOUT"`
			HealthTTL       time.Duration `conf:"default:1200ms,env:VOICE_HEALTH_TTL"`
			OutputChunkSize int32         `conf:"default:65536,env:VOICE_OUTPUT_CHUNK_SIZE"`
		}
		Auth struct {
			ArgonMemory            uint32        `conf:"default:65536"`
			ArgonIterations        uint32        `conf:"default:3"`
			ArgonParallelism       uint8         `conf:"default:2"`
			ArgonSaltLength        uint32        `conf:"default:16"`
			ArgonKeyLength         uint32        `conf:"default:32"`
			TokenHashKey           string        `conf:"noprint"`
			TokenBytes             int           `conf:"default:32"`
			SessionTTL             time.Duration `conf:"default:8h"`
			LoginRequestsPerMinute int           `conf:"default:10"`
			LoginBurst             int           `conf:"default:10"`
			LoginExpiresIn         time.Duration `conf:"default:3m"`
		}
	}{}

	const prefix = "API"

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Database Support

	db, err := database.Open(ctx, database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer db.Close()

	// -------------------------------------------------------------------------
	// SMTP Support

	smtpClient, err := mail.New(
		mail.Config{
			Host:      cfg.SMTP.Host,
			Port:      cfg.SMTP.Port,
			Username:  cfg.SMTP.Username,
			Password:  cfg.SMTP.Password,
			Timeout:   cfg.SMTP.Timeout,
			TLSPolicy: mail.TLSPolicy(cfg.SMTP.TLS),
			SSL:       cfg.SMTP.SSL,
		},
	)
	if err != nil {
		return fmt.Errorf("creating smtp client: %w", err)
	}

	// -------------------------------------------------------------------------
	// Max Adapter Support

	maxAdapterClient, err := maxadapter.New(maxadapter.Config{
		Addr:  cfg.MaxAdapter.Addr,
		Token: cfg.MaxAdapter.Token,
	})
	if err != nil {
		return fmt.Errorf("creating max adapter client: %w", err)
	}
	defer maxAdapterClient.Close()

	// -------------------------------------------------------------------------
	// Voice Adapter Support

	var voiceAdapterClient *voiceadapter.Client
	if cfg.VoiceAdapter.Addr != "" {
		voiceAdapterClient, err = voiceadapter.New(voiceadapter.Config{
			Addr:  cfg.VoiceAdapter.Addr,
			Token: cfg.VoiceAdapter.Token,
		})
		if err != nil {
			log.Error(ctx, "voice adapter disabled", "err", err)
		} else {
			defer voiceAdapterClient.Close()
		}
	}

	// -------------------------------------------------------------------------
	// Crypto and Token Support

	hasher, err := password.New(password.Argon2idConfig{
		Memory:      cfg.Auth.ArgonMemory,
		Iterations:  cfg.Auth.ArgonIterations,
		Parallelism: cfg.Auth.ArgonParallelism,
		SaltLength:  cfg.Auth.ArgonSaltLength,
		KeyLength:   cfg.Auth.ArgonKeyLength,
	})
	if err != nil {
		return fmt.Errorf("creating password hasher: %w", err)
	}

	tokenManager, err := token.New(token.Config{
		TokenBytes: cfg.Auth.TokenBytes,
		HashKey:    cfg.Auth.TokenHashKey,
	})
	if err != nil {
		return fmt.Errorf("creating tokeninzier: %w", err)
	}

	// -------------------------------------------------------------------------
	// Create Business Packages

	attachmentFileStore := filestore.New(cfg.Storage.AttachmentRootDir, cfg.Storage.AttachmentPathPrefix)

	adminStore := admindb.NewStore(db)
	adminBus := adminbus.NewBusiness(hasher, adminStore)

	sessionStore := sessiondb.NewStore(db)
	sessionBus := sessionbus.NewBusiness(sessionStore, tokenManager, cfg.Auth.SessionTTL)

	eventStore := eventdb.NewStore(db)
	eventBus := eventbus.NewBusinnes(eventStore)

	departmentStore := departmentdb.NewStore(db)
	departmentBus := departmentbus.NewBusiness(departmentStore)

	employeeStore := employeedb.NewStore(db)
	employeeBus := employeebus.NewBusiness(employeeStore)

	vtargetStore := vtargetdb.NewStore(db)
	vtargetBus := vtargetbus.NewBusiness(vtargetStore)

	targetStore := targetdb.NewStore(db)
	campaignStore := campaigndb.NewStore(db)
	targetBus := campaignbus.NewTargetBusiness(campaignStore, targetStore)

	attachmentStore := attachmentdb.NewStore(db)
	attachmentBus := attachmentbus.NewBusiness(attachmentFileStore, attachmentStore)

	organizationStore := organizationdb.NewStore(db)
	organizationBus := organizationbus.NewBusiness(organizationStore, attachmentBus)

	resolverBus := resolverbus.NewBusiness(targetBus, employeeBus, departmentBus, organizationBus, attachmentBus)

	messageStore := messagedb.NewStore(db)
	messageBus := messagebus.NewBusiness(messageStore, attachmentBus)

	landingStore := landingdb.NewStore(db)
	landingBus := landingbus.NewBusiness(landingStore, attachmentBus)

	renderBus := renderbus.NewBusiness(attachmentFileStore, resolverBus)

	authBus := authbus.NewBusiness(adminBus, sessionBus)

	maxAccountStore := maxaccountdb.NewStore(db)
	maxAccountBus := maxaccountbus.NewBusiness(maxAccountStore, maxAdapterClient)

	campaignBus := campaignbus.NewCampaignBusiness(campaignStore, targetStore, messageBus, landingBus, employeeBus, resolverBus, attachmentBus)

	deliverybus := deliverybus.NewBusiness(targetBus, campaignBus, employeeBus, messageBus, attachmentBus, smtpClient, eventBus, renderBus)
	maxDeliveryStore := maxdeliverydb.NewStore(db)
	maxDeliveryBus := maxdeliverybus.NewBusiness(targetBus, campaignBus, employeeBus, messageBus, maxAccountBus, attachmentBus, renderBus, eventBus, maxAdapterClient, maxDeliveryStore)

	visitBus := visitbus.NewBusiness(targetBus, campaignBus, landingBus, eventBus, attachmentBus, renderBus)

	// -------------------------------------------------------------------------
	// Start Workers

	workers := workers.NewWorker(log, campaignBus, deliverybus, maxDeliveryBus, sessionBus, cfg.Worker.Interval)
	workers.Run(ctx)

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	e := echo.New()

	build.Add(e, build.Config{
		Log:             log,
		OrganizationBus: organizationBus,
		DepartmentBus:   departmentBus,
		EmployeeBus:     employeeBus,
		LandingBus:      landingBus,
		MaxAccountBus:   maxAccountBus,
		MessageBus:      messageBus,
		CampaignBus:     campaignBus,
		TargetBus:       targetBus,
		VTargetBus:      vtargetBus,
		VisitBus:        visitBus,
		AttachmentBus:   attachmentBus,
		RenderBus:       renderBus,
		VoiceAdapter:    voiceAdapterClient,
		VoiceRuntime: build.VoiceRuntimeConfig{
			CloneTimeout:    cfg.Voice.CloneTimeout,
			ASRTimeout:      cfg.Voice.ASRTimeout,
			HealthTTL:       cfg.Voice.HealthTTL,
			OutputChunkSize: cfg.Voice.OutputChunkSize,
		},
		AuthBus:    authBus,
		SessionBus: sessionBus,
		LoginRateLimit: build.LoginRateLimitConfig{
			RequestsPerMinute: cfg.Auth.LoginRequestsPerMinute,
			Burst:             cfg.Auth.LoginBurst,
			ExpiresIn:         cfg.Auth.LoginExpiresIn,
		},
	})

	s := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      e,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", s.Addr)

		serverErrors <- s.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		workers.Stop(ctx)
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.App.ShutdownTimeout)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			s.Close()
			return fmt.Errorf("server shutdown: %w", err)
		}

		if err := workers.Stop(ctx); err != nil {
			log.Error(ctx, "shutdown", "status", "stopping workers", "err", err)
		}
	}

	return nil
}
