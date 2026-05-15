package sessionbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAdminNotFound = errors.New("admin not found")
	ErrInvalidCSRF   = errors.New("invalid csrf token")
	ErrExpired       = errors.New("session expired")
	ErrRevoked       = errors.New("session revoked")
)

type Storer interface {
	Save(ctx context.Context, session Session) error
	QueryByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	Revoke(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error
	RevokeByAdminID(ctx context.Context, adminID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, expiredAt time.Time) error
}

type TokenManager interface {
	NewToken() (string, error)
	HMACSHA256(value string) string
	CompareHash(rawToken string, storedHash string) bool
}

type Business struct {
	storer Storer
	tokens TokenManager
	ttl    time.Duration
}

func NewBusiness(store Storer, tokens TokenManager, ttl time.Duration) *Business {
	return &Business{storer: store, tokens: tokens, ttl: ttl}
}

func (b *Business) Create(ctx context.Context, session NewSession) (CreatedSession, error) {
	sToken, err := b.tokens.NewToken()
	if err != nil {
		return CreatedSession{}, fmt.Errorf("save: %w", err)
	}

	cfToken, err := b.tokens.NewToken()
	if err != nil {
		return CreatedSession{}, fmt.Errorf("save: %w", err)
	}

	sTokenHash := b.tokens.HMACSHA256(sToken)

	now := time.Now().UTC()
	expiresAt := now.Add(b.ttl)

	s := Session{
		ID:        uuid.New(),
		AdminID:   session.AdminID,
		TokenHash: sTokenHash,
		CSRFToken: cfToken,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		IPAddress: session.IPAddress,
		UserAgent: session.UserAgent,
	}

	err = b.storer.Save(ctx, s)
	if err != nil {
		return CreatedSession{}, fmt.Errorf("save: %w", err)
	}

	return CreatedSession{
		Token:     sToken,
		CSRFToken: cfToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (b *Business) Authenticate(ctx context.Context, rawSessionToken string) (Session, error) {
	tokenHash := b.tokens.HMACSHA256(rawSessionToken)

	sess, err := b.storer.QueryByTokenHash(ctx, tokenHash)
	if err != nil {
		return Session{}, fmt.Errorf("authenticate: %w", err)
	}

	now := time.Now().UTC()

	if sess.RevokedAt != nil {
		return Session{}, ErrRevoked
	}

	if !sess.ExpiresAt.After(now) {
		return Session{}, ErrExpired
	}

	return sess, nil
}

func (b *Business) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()

	if err := b.storer.Revoke(ctx, sessionID, now); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}

func (b *Business) RevokeByAdminID(ctx context.Context, adminID uuid.UUID) error {
	now := time.Now().UTC()

	if err := b.storer.RevokeByAdminID(ctx, adminID, now); err != nil {
		return fmt.Errorf("revokebyadminid: %w", err)
	}

	return nil
}

func (b *Business) DeleteExpired(ctx context.Context) error {
	now := time.Now().UTC()

	if err := b.storer.DeleteExpired(ctx, now); err != nil {
		return fmt.Errorf("deleteexpired: %w", err)
	}

	return nil
}

func (b *Business) ValidateCSRFToken(ctx context.Context, sess Session, rawCSRFToken string) error {
	now := time.Now().UTC()
	if sess.RevokedAt != nil {
		return ErrRevoked
	}

	if !sess.ExpiresAt.After(now) {
		return ErrExpired
	}

	if !b.tokens.CompareHash(rawCSRFToken, sess.CSRFToken) {
		return ErrInvalidCSRF
	}

	return nil
}
