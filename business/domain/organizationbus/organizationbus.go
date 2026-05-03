package organizationbus

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

var GlobalID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var (
	ErrAlreadyExists = errors.New("organization already exists")
	ErrNotFound      = errors.New("organization not found")
)

type Storer interface {
	Save(ctx context.Context, organization Organization) error
	QueryByID(ctx context.Context, id uuid.UUID) (Organization, error)
	Update(ctx context.Context, organization Organization) error
}

type FileStorage interface {
	Save(ctx context.Context, r io.Reader, ext string) (file.Path, error)
	Delete(ctx context.Context, u file.Path) error
}

type Business struct {
	storer      Storer
	fileStorage FileStorage
}

func NewBusiness(storer Storer, fileStorage FileStorage) *Business {
	return &Business{
		storer:      storer,
		fileStorage: fileStorage,
	}
}

func (b *Business) Save(ctx context.Context, organization NewOrganization) error {
	err := b.storer.Save(ctx, Organization{
		ID:         GlobalID,
		Label:      organization.Label,
		Attributes: organization.Attributes,
	})

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (b *Business) Get(ctx context.Context) (Organization, error) {
	organization, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		return Organization{}, fmt.Errorf("get: %w", err)
	}
	return organization, nil
}

func (b *Business) Update(ctx context.Context, organization Organization, up UpdateOrganization) error {
	if up.Label != nil {
		organization.Label = *up.Label
	}

	if up.Attributes != nil {
		organization.Attributes = *up.Attributes
	}

	if err := b.storer.Update(ctx, organization); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

func (b *Business) UpdateLogo(ctx context.Context, r io.Reader, ext string) (file.Path, error) {
	newPath, err := b.fileStorage.Save(ctx, r, ext)
	if err != nil {
		return file.Path{}, fmt.Errorf("updatelogo: save file: %w", err)
	}

	org, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		_ = b.fileStorage.Delete(ctx, newPath)
		return file.Path{}, fmt.Errorf("updatelogo: get org: %w", err)
	}

	oldPath := org.LogoPath
	org.LogoPath = file.NewNullPath(newPath)

	if err := b.storer.Update(ctx, org); err != nil {
		_ = b.fileStorage.Delete(ctx, newPath)
		return file.Path{}, fmt.Errorf("updatelogo: update org: %w", err)
	}

	if oldPath.Valid() && !oldPath.Path().Equal(newPath) {
		if err = b.fileStorage.Delete(ctx, oldPath.Path()); err != nil {
			return file.Path{}, fmt.Errorf("updatelogo: delete file: %w", err)
		}
	}

	return newPath, nil
}

func (b *Business) DeleteLogo(ctx context.Context) error {
	org, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		return fmt.Errorf("deletelogo: get org: %w", err)
	}

	if !org.LogoPath.Valid() {
		return nil
	}

	logoPath := org.LogoPath.Path()
	org.LogoPath = file.Null{}

	if err := b.storer.Update(ctx, org); err != nil {
		return fmt.Errorf("deletelogo: update org: %w", err)
	}

	if err := b.fileStorage.Delete(ctx, logoPath); err != nil {
		return fmt.Errorf("deletelogo: delete file: %w", err)
	}

	return nil
}
