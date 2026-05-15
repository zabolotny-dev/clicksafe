package sessionbus

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	AdminID   uuid.UUID
	TokenHash string
	CSRFToken string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
	IPAddress netip.Addr
	UserAgent string
}

type NewSession struct {
	AdminID   uuid.UUID
	IPAddress netip.Addr
	UserAgent string
}

type CreatedSession struct {
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}
