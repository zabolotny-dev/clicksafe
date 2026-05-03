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
	ErrNotFound = errors.New("organization not found")
)

type Storer interface {
	Save(ctx context.Context, organization Organization) error
	QueryByID(ctx context.Context, id uuid.UUID) (Organization, error)
	UpdateLogo(ctx context.Context, id uuid.UUID, logoPath file.Path) error
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

func (b *Business) SaveLogo(ctx context.Context, r io.Reader, ext string) (file.Path, error) {
	newPath, err := b.fileStorage.Save(ctx, r, ext)
	if err != nil {
		return file.Path{}, fmt.Errorf("savelogo: save file: %w", err)
	}

	org, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		return file.Path{}, fmt.Errorf("savelogo: get org: %w", err)
	}

	if err := b.storer.UpdateLogo(ctx, GlobalID, newPath); err != nil {
		return file.Path{}, fmt.Errorf("savelogo: update org: %w", err)
	}

	if !org.LogoPath.IsEmpty() && !org.LogoPath.Equal(newPath) {
		if err = b.fileStorage.Delete(ctx, org.LogoPath); err != nil {
			return file.Path{}, fmt.Errorf("savelogo: delete file: %w", err)
		}
	}

	return newPath, nil
}

func (b *Business) DeleteLogo(ctx context.Context) error {
	org, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		return fmt.Errorf("deletelogo: get org: %w", err)
	}

	if org.LogoPath.IsEmpty() {
		return nil
	}

	if err := b.storer.UpdateLogo(ctx, GlobalID, file.Path{}); err != nil {
		return fmt.Errorf("deletelogo: update org: %w", err)
	}

	if err := b.fileStorage.Delete(ctx, org.LogoPath); err != nil {
		return fmt.Errorf("deletelogo: delete file: %w", err)
	}

	return nil
}
