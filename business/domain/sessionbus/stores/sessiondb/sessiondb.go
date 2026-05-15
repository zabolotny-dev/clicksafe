package sessiondb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus/stores/sessiondb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
)

const adminFKConstraint = "sessions_admin_id_fkey"

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, sess sessionbus.Session) error {
	dbSess := toDBSession(sess)

	err := s.q.Save(ctx, sqlc.SaveParams{
		ID:        dbSess.ID,
		AdminID:   dbSess.AdminID,
		TokenHash: dbSess.TokenHash,
		CsrfToken: dbSess.CsrfToken,
		CreatedAt: dbSess.CreatedAt,
		ExpiresAt: dbSess.ExpiresAt,
		RevokedAt: dbSess.RevokedAt,
		IpAddress: dbSess.IpAddress,
		UserAgent: dbSess.UserAgent,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.FKViolation && pgErr.ConstraintName == adminFKConstraint {
			return sessionbus.ErrAdminNotFound
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) QueryByTokenHash(ctx context.Context, tokenHash string) (sessionbus.Session, error) {
	dbSess, err := s.q.QueryByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessionbus.Session{}, sessionbus.ErrExpired
		}
		return sessionbus.Session{}, fmt.Errorf("db: %w", err)
	}

	return toBusSession(dbSess), nil
}

func (s *Store) Revoke(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error {
	err := s.q.Revoke(ctx, sqlc.RevokeParams{
		ID:        sessionID,
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) RevokeByAdminID(ctx context.Context, adminID uuid.UUID, revokedAt time.Time) error {
	err := s.q.RevokeByAdminID(ctx, sqlc.RevokeByAdminIDParams{
		AdminID:   adminID,
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) DeleteExpired(ctx context.Context, expiredAt time.Time) error {
	if err := s.q.DeleteExpired(ctx, pgtype.Timestamptz{Time: expiredAt, Valid: true}); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
