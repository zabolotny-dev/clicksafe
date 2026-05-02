package employeedb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus/stores/employeedb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

const (
	uniqueEmailConstraint = "employees_email_key"
	uniquePhoneConstraint = "employees_phone_number_key"
)

type Store struct {
	q *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Save(ctx context.Context, emp employeebus.Employee) (employeebus.Employee, error) {
	dbEmp, err := toDBEmployee(emp)
	if err != nil {
		return employeebus.Employee{}, fmt.Errorf("db: %w", err)
	}

	err = s.q.Save(ctx, sqlc.SaveParams{
		ID:           dbEmp.ID,
		DepartmentID: dbEmp.DepartmentID,
		FirstName:    dbEmp.FirstName,
		LastName:     dbEmp.LastName,
		Email:        dbEmp.Email,
		PhoneNumber:  dbEmp.PhoneNumber,
		Attributes:   dbEmp.Attributes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			switch pgErr.ConstraintName {
			case uniqueEmailConstraint:
				return employeebus.Employee{}, employeebus.ErrUniqueEmail
			case uniquePhoneConstraint:
				return employeebus.Employee{}, employeebus.ErrUniquePhone
			}
		}
		return employeebus.Employee{}, fmt.Errorf("db: %w", err)
	}

	return emp, nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error) {
	dbEmp, err := s.q.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return employeebus.Employee{}, employeebus.ErrNotFound
		}
		return employeebus.Employee{}, fmt.Errorf("db: %w", err)
	}

	emp, err := toBusEmployee(dbEmp)
	if err != nil {
		return employeebus.Employee{}, fmt.Errorf("db: %w", err)
	}

	return emp, nil
}

func (s *Store) Query(ctx context.Context, filter employeebus.QueryFilter, orderBy order.By, page page.Page) ([]employeebus.Employee, error) {
	var fullNameFilter pgtype.Text
	if filter.FullName != nil {
		fullNameFilter = pgtype.Text{String: *filter.FullName, Valid: true}
	}

	var emailFilter pgtype.Text
	if filter.Email != nil {
		emailFilter = pgtype.Text{String: *filter.Email, Valid: true}
	}

	var phoneNumberFilter pgtype.Text
	if filter.Phone != nil {
		phoneNumberFilter = pgtype.Text{String: *filter.Phone, Valid: true}
	}

	dbEmps, err := s.q.Query(ctx, sqlc.QueryParams{
		ID:           filter.ID,
		DepartmentID: filter.DepartmentID,
		FullName:     fullNameFilter,
		Email:        emailFilter,
		PhoneNumber:  phoneNumberFilter,
		OrderBy:      orderBy.SQLOrderBy(),
		OffsetVal:    int32((page.Number() - 1) * page.RowsPerPage()),
		LimitVal:     int32(page.RowsPerPage()),
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	emps, err := toBusEmployees(dbEmps)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return emps, nil
}

func (s *Store) Update(ctx context.Context, emp employeebus.Employee) error {
	dbEmp, err := toDBEmployee(emp)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	err = s.q.Update(ctx, sqlc.UpdateParams{
		ID:           dbEmp.ID,
		DepartmentID: dbEmp.DepartmentID,
		FirstName:    dbEmp.FirstName,
		LastName:     dbEmp.LastName,
		Email:        dbEmp.Email,
		PhoneNumber:  dbEmp.PhoneNumber,
		Attributes:   dbEmp.Attributes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.UniqueViolation {
			switch pgErr.ConstraintName {
			case uniqueEmailConstraint:
				return employeebus.ErrUniqueEmail
			case uniquePhoneConstraint:
				return employeebus.ErrUniquePhone
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.q.Delete(ctx, id); err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func (s *Store) Count(ctx context.Context, filter employeebus.QueryFilter) (int, error) {
	var fullNameFilter pgtype.Text
	if filter.FullName != nil {
		fullNameFilter = pgtype.Text{String: *filter.FullName, Valid: true}
	}

	var emailFilter pgtype.Text
	if filter.Email != nil {
		emailFilter = pgtype.Text{String: *filter.Email, Valid: true}
	}

	var phoneNumberFilter pgtype.Text
	if filter.Phone != nil {
		phoneNumberFilter = pgtype.Text{String: *filter.Phone, Valid: true}
	}

	count, err := s.q.Count(ctx, sqlc.CountParams{
		ID:           filter.ID,
		DepartmentID: filter.DepartmentID,
		FullName:     fullNameFilter,
		Email:        emailFilter,
		PhoneNumber:  phoneNumberFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(count), nil
}
