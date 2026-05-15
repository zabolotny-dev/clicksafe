package authbus

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

type AdminManager interface {
	Authenticate(ctx context.Context, login login.Login, password password.Password) (adminbus.Admin, error)
	QueryByID(ctx context.Context, id uuid.UUID) (adminbus.Admin, error)
}

type SessionManager interface {
	Create(ctx context.Context, session sessionbus.NewSession) (sessionbus.CreatedSession, error)
	Revoke(ctx context.Context, sessionID uuid.UUID) error
}

type LoginData struct {
	Login     login.Login
	Password  password.Password
	IPAddress netip.Addr
	UserAgent string
}

type LoginResult struct {
	Admin   adminbus.Admin
	Session sessionbus.CreatedSession
}

type Business struct {
	adminManager   AdminManager
	sessionManager SessionManager
}

func NewBusiness(am AdminManager, sm SessionManager) *Business {
	return &Business{
		adminManager:   am,
		sessionManager: sm,
	}
}

func (b *Business) Login(ctx context.Context, ld LoginData) (LoginResult, error) {
	ad, err := b.adminManager.Authenticate(ctx, ld.Login, ld.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("login: %w", err)
	}

	sess, err := b.sessionManager.Create(ctx, sessionbus.NewSession{
		AdminID:   ad.ID,
		IPAddress: ld.IPAddress,
		UserAgent: ld.UserAgent,
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("login: %w", err)
	}

	return LoginResult{
		Admin:   ad,
		Session: sess,
	}, nil
}

func (b *Business) Logout(ctx context.Context, session sessionbus.Session) error {
	err := b.sessionManager.Revoke(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

func (b *Business) Me(ctx context.Context, session sessionbus.Session) (adminbus.Admin, error) {
	ad, err := b.adminManager.QueryByID(ctx, session.AdminID)
	if err != nil {
		return adminbus.Admin{}, fmt.Errorf("me: %w", err)
	}
	return ad, nil
}
