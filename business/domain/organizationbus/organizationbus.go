package organizationbus

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

var GlobalID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var (
	ErrNotFound = errors.New("organization not found")
)

type Storer interface {
	Save(ctx context.Context, organization Organization) error
	Get(ctx context.Context, id uuid.UUID) (Organization, error)
	UpdateLogo(ctx context.Context, id uuid.UUID, logoURL url.URL) error
}

type FileStorage interface {
	Save(ctx context.Context, r io.Reader, ext string) (url.URL, error)
	Delete(ctx context.Context, u url.URL) error
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
		Name:       organization.Name,
		Attributes: organization.Attributes,
	})

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (b *Business) Get(ctx context.Context) (Organization, error) {
	organization, err := b.storer.Get(ctx, GlobalID)
	if err != nil {
		return Organization{}, fmt.Errorf("get: %w", err)
	}
	return organization, nil
}

func (b *Business) SaveLogo(ctx context.Context, r io.Reader, ext string) (url.URL, error) {
	newURL, err := b.fileStorage.Save(ctx, r, ext)
	if err != nil {
		return url.URL{}, fmt.Errorf("savelogo: save file: %w", err)
	}

	org, err := b.storer.Get(ctx, GlobalID)
	if err != nil {
		return url.URL{}, fmt.Errorf("savelogo: get org: %w", err)
	}

	if !org.LogoURL.IsEmpty() && org.LogoURL != newURL {
		if err = b.fileStorage.Delete(ctx, org.LogoURL); err != nil {
			return url.URL{}, fmt.Errorf("savelogo: delete file: %w", err)
		}
	}

	if err := b.storer.UpdateLogo(ctx, GlobalID, newURL); err != nil {
		return url.URL{}, fmt.Errorf("savelogo: update org: %w", err)
	}

	return newURL, nil
}
