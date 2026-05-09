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
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus/stores/vtargetdb"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/usecase/deliverybus"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
	"github.com/zabolotny-dev/clicksafe/foundation/logger"
	"github.com/zabolotny-dev/clicksafe/foundation/mail"
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
			RootDir           string `conf:"default:./public"`
			PathPrefix        string `conf:"default:/uploads"`
			MessageRootDir    string `conf:"default:./private/messages"`
			MessagePathPrefix string `conf:"default:/messages"`
			LandingRootDir    string `conf:"default:./private/landings"`
			LandingPathPrefix string `conf:"default:/landings"`
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
	// Create Business Packages

	publicFileStore := filestore.New(cfg.Storage.RootDir, cfg.Storage.PathPrefix)
	messageFileStore := filestore.New(cfg.Storage.MessageRootDir, cfg.Storage.MessagePathPrefix)
	landingFileStore := filestore.New(cfg.Storage.LandingRootDir, cfg.Storage.LandingPathPrefix)

	eventStore := eventdb.NewStore(db)
	eventBus := eventbus.NewBusinnes(eventStore)

	organizationStore := organizationdb.NewStore(db)
	organizationBus := organizationbus.NewBusiness(organizationStore, publicFileStore)

	departmentStore := departmentdb.NewStore(db)
	departmentBus := departmentbus.NewBusiness(departmentStore)

	employeeStore := employeedb.NewStore(db)
	employeeBus := employeebus.NewBusiness(employeeStore)

	targetStore := targetdb.NewStore(db)
	campaignStore := campaigndb.NewStore(db)

	campaignBus := campaignbus.NewCampaignBusiness(campaignStore, targetStore)
	targetBus := campaignbus.NewTargetBusiness(campaignStore, targetStore)

	vtargetStore := vtargetdb.NewStore(db)
	vtargetBus := vtargetbus.NewBusiness(vtargetStore)

	resolverBus := resolverbus.NewBusiness(targetBus, targetBus, employeeBus, departmentBus, organizationBus)

	messageStore := messagedb.NewStore(db)
	messageBus := messagebus.NewBusiness(messageStore, messageFileStore, resolverBus)

	landingStore := landingdb.NewStore(db)
	landingBus := landingbus.NewBusiness(landingStore, landingFileStore, resolverBus)

	deliverybus := deliverybus.NewBusiness(targetBus, campaignBus, employeeBus, messageBus, smtpClient, eventBus)

	visitBus := visitbus.NewBusiness(targetBus, campaignBus, landingBus, eventBus)

	// -------------------------------------------------------------------------
	// Start Workers

	workers := workers.NewWorker(log, campaignBus, deliverybus, cfg.Worker.Interval)
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
		MessageBus:      messageBus,
		CampaignBus:     campaignBus,
		TargetBus:       targetBus,
		VTargetBus:      vtargetBus,
		VisitBus:        visitBus,
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
