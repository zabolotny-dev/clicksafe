package sessiondb

import (
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus/stores/sessiondb/sqlc"
)

func toDBSession(sess sessionbus.Session) sqlc.Session {
	var revokedAt pgtype.Timestamptz
	if sess.RevokedAt != nil {
		revokedAt = pgtype.Timestamptz{Time: *sess.RevokedAt, Valid: true}
	}

	var ipAddress *netip.Addr
	if sess.IPAddress.IsValid() {
		addr := sess.IPAddress
		ipAddress = &addr
	}

	var userAgent pgtype.Text
	if sess.UserAgent != "" {
		userAgent = pgtype.Text{String: sess.UserAgent, Valid: true}
	}

	return sqlc.Session{
		ID:        sess.ID,
		AdminID:   sess.AdminID,
		TokenHash: sess.TokenHash,
		CsrfToken: sess.CSRFToken,
		CreatedAt: pgtype.Timestamptz{Time: sess.CreatedAt, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: sess.ExpiresAt, Valid: true},
		RevokedAt: revokedAt,
		IpAddress: ipAddress,
		UserAgent: userAgent,
	}
}

func toBusSession(sess sqlc.Session) sessionbus.Session {
	var revokedAt *time.Time
	if sess.RevokedAt.Valid {
		revokedAt = &sess.RevokedAt.Time
	}

	var ipAddress netip.Addr
	if sess.IpAddress != nil {
		ipAddress = *sess.IpAddress
	}

	return sessionbus.Session{
		ID:        sess.ID,
		AdminID:   sess.AdminID,
		TokenHash: sess.TokenHash,
		CSRFToken: sess.CsrfToken,
		CreatedAt: sess.CreatedAt.Time,
		ExpiresAt: sess.ExpiresAt.Time,
		RevokedAt: revokedAt,
		IPAddress: ipAddress,
		UserAgent: sess.UserAgent.String,
	}
}
