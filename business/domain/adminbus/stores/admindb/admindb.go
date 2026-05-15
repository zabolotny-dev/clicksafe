package admindb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus/stores/admindb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
)

const uniqueLoginConstraint = "admins_login_key"

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, adm adminbus.Admin) error {
	dbAdm, err := toDBAdmin(adm)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Save(ctx, sqlc.SaveParams{
		ID:           dbAdm.ID,
		FirstName:    dbAdm.FirstName,
		LastName:     dbAdm.LastName,
		Login:        dbAdm.Login,
		PasswordHash: dbAdm.PasswordHash,
		CreatedAt:    dbAdm.CreatedAt,
	})
	if err != nil {
		return parseSaveUpdateError(err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, adm adminbus.Admin) error {
	dbAdm, err := toDBAdmin(adm)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Update(ctx, sqlc.UpdateParams{
		ID:           dbAdm.ID,
		FirstName:    dbAdm.FirstName,
		LastName:     dbAdm.LastName,
		Login:        dbAdm.Login,
		PasswordHash: dbAdm.PasswordHash,
	})
	if err != nil {
		return parseSaveUpdateError(err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (adminbus.Admin, error) {
	dbAdm, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminbus.Admin{}, adminbus.ErrNotFound
		}
		return adminbus.Admin{}, fmt.Errorf("db: %w", err)
	}

	adm, err := toBusAdmin(dbAdm)
	if err != nil {
		return adminbus.Admin{}, fmt.Errorf("db: %w", err)
	}

	return adm, nil
}

func (s *Store) QueryByLogin(ctx context.Context, login login.Login) (adminbus.Admin, error) {
	dbAdm, err := s.q.QueryByLogin(ctx, login.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminbus.Admin{}, adminbus.ErrInvalidCredential
		}
		return adminbus.Admin{}, fmt.Errorf("db: %w", err)
	}

	adm, err := toBusAdmin(dbAdm)
	if err != nil {
		return adminbus.Admin{}, fmt.Errorf("db: %w", err)
	}

	return adm, nil
}

func (s *Store) Query(ctx context.Context, filter adminbus.AdminQueryFilter) ([]adminbus.Admin, error) {
	var loginFilter pgtype.Text
	if filter.Login != nil {
		loginFilter = pgtype.Text{String: *filter.Login, Valid: true}
	}

	var fullNameFilter pgtype.Text
	if filter.FullName != nil {
		fullNameFilter = pgtype.Text{String: *filter.FullName, Valid: true}
	}

	dbAdmins, err := s.q.Query(ctx, sqlc.QueryParams{
		Login:    loginFilter,
		FullName: fullNameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	admins, err := toBusAdmins(dbAdmins)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return admins, nil
}

func (s *Store) Count(ctx context.Context, filter adminbus.AdminQueryFilter) (int, error) {
	var loginFilter pgtype.Text
	if filter.Login != nil {
		loginFilter = pgtype.Text{String: *filter.Login, Valid: true}
	}

	var fullNameFilter pgtype.Text
	if filter.FullName != nil {
		fullNameFilter = pgtype.Text{String: *filter.FullName, Valid: true}
	}

	count, err := s.q.Count(ctx, sqlc.CountParams{
		Login:    loginFilter,
		FullName: fullNameFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(count), nil
}

func parseSaveUpdateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation && pgErr.ConstraintName == uniqueLoginConstraint {
		return adminbus.ErrUniqueLogin
	}

	return fmt.Errorf("db: %w", err)
}
