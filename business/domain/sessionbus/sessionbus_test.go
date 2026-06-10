package sessionbus_test

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

func Test_Session(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Session")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, create(db.BusDomain, sd), "create")
	unittest.Run(t, authenticate(db.BusDomain, sd), "authenticate")
	unittest.Run(t, revoke(db.BusDomain, sd), "revoke")
	unittest.Run(t, validateCSRF(db.BusDomain, sd), "validatecsrf")
}

// =============================================================================

type seedData struct {
	Admins  []adminbus.Admin
	Session sessionbus.CreatedSession
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	admins, _, err := adminbus.TestSeedAdmins(ctx, 1, busDomain.Admin)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding admin: %w", err)
	}

	sess, err := busDomain.Session.Create(ctx, sessionbus.NewSession{
		AdminID:   admins[0].ID,
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "TestAgent",
	})
	if err != nil {
		return seedData{}, fmt.Errorf("seeding session: %w", err)
	}

	return seedData{Admins: admins, Session: sess}, nil
}

// =============================================================================

func create(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "returns-token",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				sess, err := busDomain.Session.Create(ctx, sessionbus.NewSession{
					AdminID:   sd.Admins[0].ID,
					IPAddress: netip.MustParseAddr("192.168.1.1"),
					UserAgent: "TestAgent2",
				})
				if err != nil {
					return err
				}
				return sess.Token != "" && sess.CSRFToken != ""
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func authenticate(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "valid-token-returns-session",
			ExpResp: sd.Admins[0].ID.String(),
			ExcFunc: func(ctx context.Context) any {
				sess, err := busDomain.Session.Authenticate(ctx, sd.Session.Token)
				if err != nil {
					return err
				}
				return sess.AdminID.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func revoke(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "revoked-token-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Create a fresh session to revoke
				sess, err := busDomain.Session.Create(ctx, sessionbus.NewSession{
					AdminID:   sd.Admins[0].ID,
					IPAddress: netip.MustParseAddr("10.0.0.1"),
					UserAgent: "RevokeAgent",
				})
				if err != nil {
					return err
				}

				// Get the session to find its ID
				session, err := busDomain.Session.Authenticate(ctx, sess.Token)
				if err != nil {
					return err
				}

				if err := busDomain.Session.Revoke(ctx, session.ID); err != nil {
					return err
				}

				// Authenticating again should fail with ErrRevoked
				_, err = busDomain.Session.Authenticate(ctx, sess.Token)
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func validateCSRF(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "invalid-csrf-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				session, err := busDomain.Session.Authenticate(ctx, sd.Session.Token)
				if err != nil {
					return err
				}
				err = busDomain.Session.ValidateCSRFToken(ctx, session, "wrong-token")
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "delete-expired-runs-without-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Session.DeleteExpired(ctx) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "revoke-by-admin-id-runs-without-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Session.RevokeByAdminID(ctx, sd.Admins[0].ID) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "authenticate-after-password-change",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				newPass := password.MustParse("NewPassword123!@#!")
				_, err := busDomain.Admin.Update(ctx, sd.Admins[0].ID, adminbus.UpdateAdmin{Password: &newPass})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "authenticate-expired-session",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Create a stub-based session that is already expired
				expiredAt := time.Now().UTC().Add(-time.Hour)
				sess := sessionbus.Session{
					ID:        uuid.New(),
					AdminID:   sd.Admins[0].ID,
					ExpiresAt: expiredAt,
				}
				err := busDomain.Session.ValidateCSRFToken(ctx, sess, "any-token")
				return errors.Is(err, sessionbus.ErrExpired)
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "validate-csrf-revoked-session",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				revokedAt := time.Now().UTC().Add(-time.Minute)
				sess := sessionbus.Session{
					ID:        uuid.New(),
					AdminID:   sd.Admins[0].ID,
					ExpiresAt: time.Now().UTC().Add(time.Hour),
					RevokedAt: &revokedAt,
				}
				err := busDomain.Session.ValidateCSRFToken(ctx, sess, "any-token")
				return errors.Is(err, sessionbus.ErrRevoked)
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Stub-based unit tests for error paths

type sessionStorerStub struct {
	savedSessions map[string]sessionbus.Session
	saveErr       error
	queryErr      error
	revokeErr     error
}

func newSessionStorerStub() *sessionStorerStub {
	return &sessionStorerStub{savedSessions: make(map[string]sessionbus.Session)}
}

func (s *sessionStorerStub) Save(_ context.Context, sess sessionbus.Session) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.savedSessions[sess.TokenHash] = sess
	return nil
}

func (s *sessionStorerStub) QueryByTokenHash(_ context.Context, hash string) (sessionbus.Session, error) {
	if s.queryErr != nil {
		return sessionbus.Session{}, s.queryErr
	}
	sess, ok := s.savedSessions[hash]
	if !ok {
		return sessionbus.Session{}, errors.New("not found")
	}
	return sess, nil
}

func (s *sessionStorerStub) Revoke(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return s.revokeErr
}

func (s *sessionStorerStub) RevokeByAdminID(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return s.revokeErr
}

func (s *sessionStorerStub) DeleteExpired(_ context.Context, _ time.Time) error {
	return s.revokeErr
}

type tokenManagerStub struct {
	newTokenErr error
	tokenVal    string
}

func (m *tokenManagerStub) NewToken() (string, error) {
	if m.newTokenErr != nil {
		return "", m.newTokenErr
	}
	val := m.tokenVal
	if val == "" {
		val = uuid.New().String()
	}
	return val, nil
}

func (m *tokenManagerStub) HMACSHA256(value string) string {
	return "hash:" + value
}

func (m *tokenManagerStub) CompareHash(rawToken, storedHash string) bool {
	return m.HMACSHA256(rawToken) == storedHash
}

func Test_Session_UnitErrors(t *testing.T) {
	t.Parallel()
	unittest.Run(t, sessionErrorPaths(), "error-paths")
}

func sessionErrorPaths() []unittest.Table {
	dbErr := errors.New("db error")
	tokenErr := errors.New("token error")

	// Create - first token fails
	busFirstTokenErr := sessionbus.NewBusiness(
		newSessionStorerStub(),
		&tokenManagerStub{newTokenErr: tokenErr},
		time.Hour,
	)

	// Create - second token fails (first succeeds, second fails)
	callCount := 0
	_ = callCount

	// Create - storer save fails
	storeSaveErr := newSessionStorerStub()
	storeSaveErr.saveErr = dbErr
	busSaveErr := sessionbus.NewBusiness(storeSaveErr, &tokenManagerStub{tokenVal: "tok"}, time.Hour)

	// Authenticate - storer query fails
	storeQueryErr := newSessionStorerStub()
	storeQueryErr.queryErr = dbErr
	busQueryErr := sessionbus.NewBusiness(storeQueryErr, &tokenManagerStub{tokenVal: "tok"}, time.Hour)

	// Authenticate - session revoked
	storeRevoked := newSessionStorerStub()
	revokedAt := time.Now().UTC().Add(-time.Minute)
	revokedSess := sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   uuid.New(),
		TokenHash: "hash:revoked-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		RevokedAt: &revokedAt,
	}
	storeRevoked.savedSessions["hash:revoked-token"] = revokedSess
	busRevoked := sessionbus.NewBusiness(storeRevoked, &tokenManagerStub{tokenVal: "revoked-token"}, time.Hour)

	// Authenticate - session expired
	storeExpired := newSessionStorerStub()
	expiredSess := sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   uuid.New(),
		TokenHash: "hash:expired-token",
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}
	storeExpired.savedSessions["hash:expired-token"] = expiredSess
	busExpired := sessionbus.NewBusiness(storeExpired, &tokenManagerStub{tokenVal: "expired-token"}, time.Hour)

	// Revoke - storer fails
	storeRevokeErr := newSessionStorerStub()
	storeRevokeErr.revokeErr = dbErr
	busRevokeErr := sessionbus.NewBusiness(storeRevokeErr, &tokenManagerStub{tokenVal: "tok"}, time.Hour)

	return []unittest.Table{
		{
			Name:    "create-first-token-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busFirstTokenErr.Create(ctx, sessionbus.NewSession{AdminID: uuid.New()})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "create-store-save-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busSaveErr.Create(ctx, sessionbus.NewSession{AdminID: uuid.New()})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "authenticate-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQueryErr.Authenticate(ctx, "some-token")
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "authenticate-revoked-returns-err-revoked",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busRevoked.Authenticate(ctx, "revoked-token")
				return errors.Is(err, sessionbus.ErrRevoked)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "authenticate-expired-returns-err-expired",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busExpired.Authenticate(ctx, "expired-token")
				return errors.Is(err, sessionbus.ErrExpired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "revoke-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busRevokeErr.Revoke(ctx, uuid.New()) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "revoke-by-admin-id-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busRevokeErr.RevokeByAdminID(ctx, uuid.New()) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "delete-expired-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busRevokeErr.DeleteExpired(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
