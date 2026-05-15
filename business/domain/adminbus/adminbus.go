package adminbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrUniqueLogin       = errors.New("login already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
)

type Storer interface {
	Save(ctx context.Context, adm Admin) error
	Update(ctx context.Context, adm Admin) error
	QueryByID(ctx context.Context, id uuid.UUID) (Admin, error)
	QueryByLogin(ctx context.Context, login login.Login) (Admin, error)
	Query(ctx context.Context, filter AdminQueryFilter) ([]Admin, error)
	Count(ctx context.Context, filter AdminQueryFilter) (int, error)
}

type Hasher interface {
	Generate(password string) (string, error)
	Compare(password, encodedHash string) (bool, error)
}

type Business struct {
	hasher Hasher
	store  Storer
}

func NewBusiness(hasher Hasher, store Storer) *Business {
	return &Business{
		hasher: hasher,
		store:  store,
	}
}

func (b *Business) Save(ctx context.Context, ad NewAdmin) (Admin, error) {
	pass, err := b.hasher.Generate(ad.Password.String())
	if err != nil {
		return Admin{}, fmt.Errorf("save: %w", err)
	}

	admin := Admin{
		ID:           uuid.New(),
		FirstName:    ad.FirstName,
		LastName:     ad.LastName,
		Login:        ad.Login,
		PasswordHash: pass,
		CreatedAt:    time.Now().UTC(),
	}

	if err := b.store.Save(ctx, admin); err != nil {
		return Admin{}, fmt.Errorf("save: %w", err)
	}

	return admin, nil
}

func (b *Business) Update(ctx context.Context, id uuid.UUID, upd UpdateAdmin) (Admin, error) {
	adm, err := b.store.QueryByID(ctx, id)
	if err != nil {
		return Admin{}, fmt.Errorf("update: %w", err)
	}

	if upd.FirstName != nil {
		adm.FirstName = *upd.FirstName
	}
	if upd.LastName != nil {
		adm.LastName = *upd.LastName
	}
	if upd.Login != nil {
		adm.Login = *upd.Login
	}
	if upd.Password != nil {
		pass, err := b.hasher.Generate(upd.Password.String())
		if err != nil {
			return Admin{}, fmt.Errorf("update: %w", err)
		}
		adm.PasswordHash = pass
	}

	if err := b.store.Update(ctx, adm); err != nil {
		return Admin{}, fmt.Errorf("update: %w", err)
	}
	return adm, nil
}

func (b *Business) Query(ctx context.Context, filter AdminQueryFilter) ([]Admin, error) {
	admin, err := b.store.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return admin, nil
}

func (b *Business) Count(ctx context.Context, filter AdminQueryFilter) (int, error) {
	count, err := b.store.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Admin, error) {
	admin, err := b.store.QueryByID(ctx, id)
	if err != nil {
		return Admin{}, fmt.Errorf("querybyid: %w", err)
	}
	return admin, nil
}

func (b *Business) QueryByLogin(ctx context.Context, login login.Login) (Admin, error) {
	admin, err := b.store.QueryByLogin(ctx, login)
	if err != nil {
		return Admin{}, fmt.Errorf("querybylogin: %w", err)
	}
	return admin, nil
}

func (b *Business) Authenticate(ctx context.Context, login login.Login, password password.Password) (Admin, error) {
	adm, err := b.store.QueryByLogin(ctx, login)
	if err != nil {
		return Admin{}, fmt.Errorf("authenticate: %w", err)
	}

	ok, err := b.hasher.Compare(password.String(), adm.PasswordHash)
	if err != nil {
		return Admin{}, fmt.Errorf("authenticate: %w", err)
	}

	if !ok {
		return Admin{}, ErrInvalidCredential
	}

	return adm, nil
}
