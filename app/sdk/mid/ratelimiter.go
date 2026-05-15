package mid

import (
	"errors"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
)

type LoginRateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	ExpiresIn         time.Duration
}

func RateLimitLogin(cfg LoginRateLimitConfig) echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      float64(cfg.RequestsPerMinute) / 60.0,
			Burst:     cfg.Burst,
			ExpiresIn: cfg.ExpiresIn,
		},
	)

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return errs.New(errs.TooManyRequests, errors.New("rate limit exceeded"))
		},
	})
}
