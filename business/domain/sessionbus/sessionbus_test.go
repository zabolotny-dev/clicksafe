package sessionbus_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

// =============================================================================
// Stubs

type sessionStorerStub struct {
	saved   sessionbus.Session
	data    map[string]sessionbus.Session // keyed by TokenHash
	saveErr error
}

func newSessionStorerStub() *sessionStorerStub {
	return &sessionStorerStub{data: make(map[string]sessionbus.Session)}
}

func (s *sessionStorerStub) Save(_ context.Context, sess sessionbus.Session) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = sess
	s.data[sess.TokenHash] = sess
	return nil
}

func (s *sessionStorerStub) QueryByTokenHash(_ context.Context, hash string) (sessionbus.Session, error) {
	sess, ok := s.data[hash]
	if !ok {
		return sessionbus.Session{}, errors.New("session not found")
	}
	return sess, nil
}

func (s *sessionStorerStub) Revoke(_ context.Context, id uuid.UUID, revokedAt time.Time) error {
	for hash, sess := range s.data {
		if sess.ID == id {
			sess.RevokedAt = &revokedAt
			s.data[hash] = sess
			return nil
		}
	}
	return nil
}

func (s *sessionStorerStub) RevokeByAdminID(_ context.Context, adminID uuid.UUID, revokedAt time.Time) error {
	for hash, sess := range s.data {
		if sess.AdminID == adminID {
			sess.RevokedAt = &revokedAt
			s.data[hash] = sess
		}
	}
	return nil
}

func (s *sessionStorerStub) DeleteExpired(_ context.Context, _ time.Time) error {
	return nil
}

// tokenManagerStub: token=<prefix><n>, hash=hmac(<value>), compare=hmac(raw)==stored
type tokenManagerStub struct {
	counter int
	tokens  []string
}

func (m *tokenManagerStub) NewToken() (string, error) {
	m.counter++
	tok := uuid.New().String()
	m.tokens = append(m.tokens, tok)
	return tok, nil
}

func (m *tokenManagerStub) HMACSHA256(value string) string {
	return "hmac:" + value
}

func (m *tokenManagerStub) CompareHash(rawToken string, storedHash string) bool {
	return storedHash == rawToken
}

// =============================================================================
// Create tests

func TestCreate_SavesSessionAndReturnsTokens(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(store, tm, time.Hour)

	adminID := uuid.New()
	cs, err := bus.Create(context.Background(), sessionbus.NewSession{
		AdminID:   adminID,
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if cs.Token == "" {
		t.Error("Create must return a non-empty session token")
	}
	if cs.CSRFToken == "" {
		t.Error("Create must return a non-empty CSRF token")
	}
	if cs.ExpiresAt.IsZero() {
		t.Error("Create must return a non-zero expiry time")
	}

	// Сессия должна быть сохранена в стабе
	if store.saved.AdminID != adminID {
		t.Errorf("saved session AdminID = %v, want %v", store.saved.AdminID, adminID)
	}
	if store.saved.TokenHash == "" {
		t.Error("saved session must have non-empty TokenHash")
	}
}

func TestCreate_TokenHashesDiffer(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(store, tm, time.Hour)

	_, _ = bus.Create(context.Background(), sessionbus.NewSession{
		AdminID:   uuid.New(),
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "agent",
	})

	// TokenHash должен быть HMAC от сессионного токена, а не от CSRF
	if store.saved.TokenHash == store.saved.CSRFToken {
		t.Error("TokenHash and CSRFToken must not be identical")
	}
}

func TestCreate_TTLSetsExpiry(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	ttl := 2 * time.Hour
	bus := sessionbus.NewBusiness(store, tm, ttl)

	before := time.Now().Add(ttl - time.Minute)
	after := time.Now().Add(ttl + time.Minute)

	cs, err := bus.Create(context.Background(), sessionbus.NewSession{
		AdminID:   uuid.New(),
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "agent",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if cs.ExpiresAt.Before(before) || cs.ExpiresAt.After(after) {
		t.Errorf("ExpiresAt = %v, want within [%v, %v]", cs.ExpiresAt, before, after)
	}
}

func TestCreate_StoreError_Propagates(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	store.saveErr = errors.New("db error")
	bus := sessionbus.NewBusiness(store, &tokenManagerStub{}, time.Hour)

	_, err := bus.Create(context.Background(), sessionbus.NewSession{
		AdminID:   uuid.New(),
		IPAddress: netip.MustParseAddr("127.0.0.1"),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// Authenticate tests

func TestAuthenticate_ValidSession(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(store, tm, time.Hour)

	rawToken := "my-raw-token"
	tokenHash := tm.HMACSHA256(rawToken)

	sess := sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	store.data[tokenHash] = sess

	got, err := bus.Authenticate(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("Session ID = %v, want %v", got.ID, sess.ID)
	}
}

func TestAuthenticate_ExpiredSession_ReturnsErrExpired(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(store, tm, time.Hour)

	rawToken := "expired-token"
	tokenHash := tm.HMACSHA256(rawToken)

	store.data[tokenHash] = sessionbus.Session{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-time.Minute), // уже истекла
	}

	_, err := bus.Authenticate(context.Background(), rawToken)
	if !errors.Is(err, sessionbus.ErrExpired) {
		t.Fatalf("Authenticate error = %v, want %v", err, sessionbus.ErrExpired)
	}
}

func TestAuthenticate_RevokedSession_ReturnsErrRevoked(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(store, tm, time.Hour)

	rawToken := "revoked-token"
	tokenHash := tm.HMACSHA256(rawToken)
	revokedAt := time.Now().Add(-time.Minute)

	store.data[tokenHash] = sessionbus.Session{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &revokedAt,
	}

	_, err := bus.Authenticate(context.Background(), rawToken)
	if !errors.Is(err, sessionbus.ErrRevoked) {
		t.Fatalf("Authenticate error = %v, want %v", err, sessionbus.ErrRevoked)
	}
}

func TestAuthenticate_SessionNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := sessionbus.NewBusiness(newSessionStorerStub(), &tokenManagerStub{}, time.Hour)

	_, err := bus.Authenticate(context.Background(), "unknown-token")
	if err == nil {
		t.Fatal("expected error for unknown session token, got nil")
	}
}

// =============================================================================
// Revoke tests

func TestRevoke_MarksSessionAsRevoked(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	bus := sessionbus.NewBusiness(store, &tokenManagerStub{}, time.Hour)

	sessID := uuid.New()
	store.data["some-hash"] = sessionbus.Session{
		ID:        sessID,
		TokenHash: "some-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := bus.Revoke(context.Background(), sessID); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}

	sess := store.data["some-hash"]
	if sess.RevokedAt == nil {
		t.Error("Revoke must set RevokedAt on the session")
	}
}

// =============================================================================
// RevokeByAdminID tests

func TestRevokeByAdminID_RevokesAllSessions(t *testing.T) {
	t.Parallel()

	store := newSessionStorerStub()
	bus := sessionbus.NewBusiness(store, &tokenManagerStub{}, time.Hour)

	adminID := uuid.New()
	store.data["hash-1"] = sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   adminID,
		TokenHash: "hash-1",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	store.data["hash-2"] = sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   adminID,
		TokenHash: "hash-2",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	// Другой admin
	store.data["hash-3"] = sessionbus.Session{
		ID:        uuid.New(),
		AdminID:   uuid.New(),
		TokenHash: "hash-3",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := bus.RevokeByAdminID(context.Background(), adminID); err != nil {
		t.Fatalf("RevokeByAdminID returned error: %v", err)
	}

	for _, hash := range []string{"hash-1", "hash-2"} {
		if store.data[hash].RevokedAt == nil {
			t.Errorf("session %s should be revoked", hash)
		}
	}
	if store.data["hash-3"].RevokedAt != nil {
		t.Error("session for other admin must not be revoked")
	}
}

// =============================================================================
// ValidateCSRFToken tests

func TestValidateCSRFToken_Valid(t *testing.T) {
	t.Parallel()

	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(newSessionStorerStub(), tm, time.Hour)

	rawCSRF := "raw-csrf-token"
	sess := sessionbus.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
		CSRFToken: rawCSRF, // CompareHash(raw, stored) => stored == raw
	}

	if err := bus.ValidateCSRFToken(context.Background(), sess, rawCSRF); err != nil {
		t.Fatalf("ValidateCSRFToken returned error: %v", err)
	}
}

func TestValidateCSRFToken_Invalid_ReturnsErrInvalidCSRF(t *testing.T) {
	t.Parallel()

	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(newSessionStorerStub(), tm, time.Hour)

	sess := sessionbus.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
		CSRFToken: "correct-csrf",
	}

	_, err := func() (interface{}, error) {
		return nil, bus.ValidateCSRFToken(context.Background(), sess, "wrong-csrf")
	}()
	if !errors.Is(err, sessionbus.ErrInvalidCSRF) {
		t.Fatalf("ValidateCSRFToken error = %v, want %v", err, sessionbus.ErrInvalidCSRF)
	}
}

func TestValidateCSRFToken_ExpiredSession_ReturnsErrExpired(t *testing.T) {
	t.Parallel()

	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(newSessionStorerStub(), tm, time.Hour)

	sess := sessionbus.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-time.Minute), // истекла
		CSRFToken: "csrf",
	}

	err := bus.ValidateCSRFToken(context.Background(), sess, "csrf")
	if !errors.Is(err, sessionbus.ErrExpired) {
		t.Fatalf("ValidateCSRFToken error = %v, want %v", err, sessionbus.ErrExpired)
	}
}

func TestValidateCSRFToken_RevokedSession_ReturnsErrRevoked(t *testing.T) {
	t.Parallel()

	tm := &tokenManagerStub{}
	bus := sessionbus.NewBusiness(newSessionStorerStub(), tm, time.Hour)

	revokedAt := time.Now().Add(-time.Minute)
	sess := sessionbus.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &revokedAt,
		CSRFToken: "csrf",
	}

	err := bus.ValidateCSRFToken(context.Background(), sess, "csrf")
	if !errors.Is(err, sessionbus.ErrRevoked) {
		t.Fatalf("ValidateCSRFToken error = %v, want %v", err, sessionbus.ErrRevoked)
	}
}
