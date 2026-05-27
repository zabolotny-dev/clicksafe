package organizationbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
)

var GlobalID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var (
	ErrAlreadyExists     = errors.New("organization already exists")
	ErrNotFound          = errors.New("organization not found")
	ErrInvalidAttachment = errors.New("invalid attachment")
)

type Storer interface {
	Save(ctx context.Context, organization Organization) error
	QueryByID(ctx context.Context, id uuid.UUID) (Organization, error)
	Update(ctx context.Context, organization Organization) error
}

type AttachmentQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type Business struct {
	storer      Storer
	attachments AttachmentQuerier
}

func NewBusiness(storer Storer, attachments AttachmentQuerier) *Business {
	return &Business{
		storer:      storer,
		attachments: attachments,
	}
}

func (b *Business) Save(ctx context.Context, organization NewOrganization) (Organization, error) {
	org := Organization{
		ID:         GlobalID,
		Label:      organization.Label,
		Attributes: organization.Attributes,
	}

	if organization.AttachmentID.Valid {
		atch, err := b.attachments.QueryByID(ctx, organization.AttachmentID.UUID)
		if err != nil {
			return Organization{}, fmt.Errorf("save: %w", err)
		}

		if !atch.Type.IsImage() {
			return Organization{}, fmt.Errorf("save: %w", ErrInvalidAttachment)
		}

		org.AttachmentID = organization.AttachmentID
	}

	if err := b.storer.Save(ctx, org); err != nil {
		return Organization{}, fmt.Errorf("save: %w", err)
	}

	return org, nil
}

func (b *Business) Get(ctx context.Context) (Organization, error) {
	organization, err := b.storer.QueryByID(ctx, GlobalID)
	if err != nil {
		return Organization{}, fmt.Errorf("get: %w", err)
	}
	return organization, nil
}

func (b *Business) Update(ctx context.Context, organization Organization, up UpdateOrganization) (Organization, error) {
	if up.Label != nil {
		organization.Label = *up.Label
	}

	if up.Attributes != nil {
		organization.Attributes = *up.Attributes
	}

	if up.AttachmentID != nil {
		if up.AttachmentID.Valid {
			atch, err := b.attachments.QueryByID(ctx, up.AttachmentID.UUID)
			if err != nil {
				return Organization{}, fmt.Errorf("update: %w", err)
			}

			if !atch.Type.IsImage() {
				return Organization{}, fmt.Errorf("update: %w", ErrInvalidAttachment)
			}
		}

		organization.AttachmentID = *up.AttachmentID
	}

	if err := b.storer.Update(ctx, organization); err != nil {
		return Organization{}, fmt.Errorf("update: %w", err)
	}

	return organization, nil
}
