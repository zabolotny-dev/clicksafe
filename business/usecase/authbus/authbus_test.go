package authbus_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
	"github.com/zabolotny-dev/clicksafe/business/usecase/authbus"
)

// =============================================================================
// Mocks

type adminManagerMock struct {
	admin    adminbus.Admin
	authErr  error
	queryErr error
}

func (m *adminManagerMock) Authenticate(_ context.Context, _ login.Login, _ password.Password) (adminbus.Admin, error) {
	return m.admin, m.authErr
}

func (m *adminManagerMock) QueryByID(_ context.Context, _ uuid.UUID) (adminbus.Admin, error) {
	return m.admin, m.queryErr
}

type sessionManagerMock struct {
	created    sessionbus.CreatedSession
	createErr  error
	revokeErr  error
	revokedIDs []uuid.UUID
}

func (m *sessionManagerMock) Create(_ context.Context, _ sessionbus.NewSession) (sessionbus.CreatedSession, error) {
	return m.created, m.createErr
}

func (m *sessionManagerMock) Revoke(_ context.Context, sessionID uuid.UUID) error {
	m.revokedIDs = append(m.revokedIDs, sessionID)
	return m.revokeErr
}

// =============================================================================

func Test_Auth(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testLogin(), "login")
	unittest.Run(t, testLogout(), "logout")
	unittest.Run(t, testMe(), "me")
}

// =============================================================================

func testLogin() []unittest.Table {
	lgn := login.MustParse("test.user")
	pass := password.MustParse("Pass123456789!")
	ip := netip.MustParseAddr("127.0.0.1")

	admin := adminbus.Admin{ID: uuid.New(), Login: lgn}
	sess := sessionbus.CreatedSession{Token: "tok-abc"}

	busOK := authbus.NewBusiness(
		&adminManagerMock{admin: admin},
		&sessionManagerMock{created: sess},
	)

	busAuthFail := authbus.NewBusiness(
		&adminManagerMock{authErr: adminbus.ErrInvalidCredential},
		&sessionManagerMock{},
	)

	busSessionFail := authbus.NewBusiness(
		&adminManagerMock{admin: admin},
		&sessionManagerMock{createErr: errors.New("session store error")},
	)

	ld := authbus.LoginData{Login: lgn, Password: pass, IPAddress: ip, UserAgent: "TestAgent"}

	return []unittest.Table{
		{
			Name:    "success",
			ExpResp: admin.ID,
			ExcFunc: func(ctx context.Context) any {
				result, err := busOK.Login(ctx, ld)
				if err != nil {
					return err
				}
				return result.Admin.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "returns-session-token",
			ExpResp: sess.Token,
			ExcFunc: func(ctx context.Context) any {
				result, err := busOK.Login(ctx, ld)
				if err != nil {
					return err
				}
				return result.Session.Token
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "auth-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busAuthFail.Login(ctx, ld)
				return errors.Is(err, adminbus.ErrInvalidCredential)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "session-create-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busSessionFail.Login(ctx, ld)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testLogout() []unittest.Table {
	sessID := uuid.New()

	sm := &sessionManagerMock{}
	bus := authbus.NewBusiness(&adminManagerMock{}, sm)

	smErr := &sessionManagerMock{revokeErr: errors.New("revoke error")}
	busErr := authbus.NewBusiness(&adminManagerMock{}, smErr)

	sess := sessionbus.Session{ID: sessID}

	return []unittest.Table{
		{
			Name:    "success-revokes-session",
			ExpResp: sessID,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.Logout(ctx, sess); err != nil {
					return err
				}
				if len(sm.revokedIDs) == 0 {
					return uuid.UUID{}
				}
				return sm.revokedIDs[0]
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "revoke-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busErr.Logout(ctx, sess)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testMe() []unittest.Table {
	admin := adminbus.Admin{ID: uuid.New(), Login: login.MustParse("find.me")}
	sess := sessionbus.Session{AdminID: admin.ID}

	busOK := authbus.NewBusiness(
		&adminManagerMock{admin: admin},
		&sessionManagerMock{},
	)

	busErr := authbus.NewBusiness(
		&adminManagerMock{queryErr: adminbus.ErrNotFound},
		&sessionManagerMock{},
	)

	return []unittest.Table{
		{
			Name:    "success",
			ExpResp: admin.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busOK.Me(ctx, sess)
				if err != nil {
					return err
				}
				return got.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busErr.Me(ctx, sess)
				return errors.Is(err, adminbus.ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
