package voiceapp

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/mid"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/voiceadapter"
)

type RuntimeConfig struct {
	CloneTimeout    time.Duration
	ASRTimeout      time.Duration
	HealthTTL       time.Duration
	OutputChunkSize int32
}

type Config struct {
	VoiceAdapterClient *voiceadapter.Client
	SessionBus         *sessionbus.Business
	Runtime            RuntimeConfig
}

func Routes(router *echo.Echo, cfg Config) {
	api := newApp(cfg.VoiceAdapterClient, cfg.Runtime)

	authen := mid.Authenticate(cfg.SessionBus)
	csrf := mid.CSRF()

	router.GET("/voice/status", api.status, authen)
	router.POST("/voice/transcribe", api.transcribe, authen, csrf)
	router.POST("/voice/clone", api.clone, authen, csrf)
}
