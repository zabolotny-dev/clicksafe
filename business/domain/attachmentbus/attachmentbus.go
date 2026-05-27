package attachmentbus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

type Storer interface {
	Save(ctx context.Context, atch Attachment) error
	Update(ctx context.Context, atch Attachment) error
	Delete(ctx context.Context, atch Attachment) error
	QueryByID(ctx context.Context, id uuid.UUID) (Attachment, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Attachment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type FileStorage interface {
	Save(ctx context.Context, r io.Reader, ext string) (file.Path, error)
	Read(ctx context.Context, p file.Path) ([]byte, error)
	Open(ctx context.Context, p file.Path) (io.ReadCloser, error)
	Delete(ctx context.Context, p file.Path) error
}

type Business struct {
	fileStore FileStorage
	storer    Storer
}

func NewBusiness(fileStore FileStorage, storer Storer) *Business {
	return &Business{fileStore: fileStore, storer: storer}
}

func (b *Business) Save(ctx context.Context, atch NewAttachment) (Attachment, error) {
	content, err := io.ReadAll(atch.Content)
	if err != nil {
		return Attachment{}, fmt.Errorf("save: %w", err)
	}

	if len(content) == 0 {
		return Attachment{}, ErrEmptyContent
	}

	var vars []string
	if atch.Type.IsTemplate() {
		if len(bytes.TrimSpace(content)) == 0 {
			return Attachment{}, ErrEmptyContent
		}

		vars, err = validateAndExtractRequiredVars(content, atch.Type)
		if err != nil {
			return Attachment{}, fmt.Errorf("save: %w", err)
		}
	} else if atch.Type.IsMedia() {
		vars = []string{}
	} else {
		return Attachment{}, fmt.Errorf("save: %w", ErrInvalidType)
	}

	filePath, err := b.fileStore.Save(ctx, bytes.NewReader(content), atch.Type.String())
	if err != nil {
		return Attachment{}, fmt.Errorf("save: %w", err)
	}

	a := Attachment{
		ID:           uuid.New(),
		Label:        atch.Label,
		Type:         atch.Type,
		ContentPath:  filePath,
		RequiredVars: vars,
		Public:       atch.Public,
		UploadedAt:   time.Now().UTC(),
	}
	if err := b.storer.Save(ctx, a); err != nil {
		return Attachment{}, fmt.Errorf("save: %w", err)
	}
	return a, nil
}

func (b *Business) Update(ctx context.Context, atch Attachment, up UpdateAttachment) (Attachment, error) {
	if up.Label != nil {
		atch.Label = *up.Label
	}
	if up.Public != nil {
		atch.Public = *up.Public
	}

	if err := b.storer.Update(ctx, atch); err != nil {
		return Attachment{}, fmt.Errorf("update: %w", err)
	}
	return atch, nil
}

func (b *Business) UpdateContent(ctx context.Context, atch Attachment, up UpdateAttachmentContent) (Attachment, error) {
	if atch.Type != Html && atch.Type != Txt {
		return Attachment{}, fmt.Errorf("updatecontent: %w", ErrInvalidType)
	}

	if up.Content == nil {
		return Attachment{}, fmt.Errorf("updatecontent: %w", ErrEmptyContent)
	}

	content, err := io.ReadAll(up.Content)
	if err != nil {
		return Attachment{}, fmt.Errorf("updatecontent: %w", err)
	}

	if len(bytes.TrimSpace(content)) == 0 {
		return Attachment{}, fmt.Errorf("updatecontent: %w", ErrEmptyContent)
	}

	vars, err := validateAndExtractRequiredVars(content, atch.Type)
	if err != nil {
		return Attachment{}, fmt.Errorf("updatecontent: %w", err)
	}

	oldPath := atch.ContentPath
	newPath, err := b.fileStore.Save(ctx, bytes.NewReader(content), atch.Type.String())
	if err != nil {
		return Attachment{}, fmt.Errorf("updatecontent: %w", err)
	}

	atch.ContentPath = newPath
	atch.RequiredVars = vars
	atch.UploadedAt = time.Now().UTC()

	if err := b.storer.Update(ctx, atch); err != nil {
		_ = b.fileStore.Delete(ctx, newPath)
		return Attachment{}, fmt.Errorf("updatecontent: %w", err)
	}

	_ = b.fileStore.Delete(ctx, oldPath)

	return atch, nil
}

func (b *Business) Delete(ctx context.Context, atch Attachment) error {
	if err := b.storer.Delete(ctx, atch); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if err := b.fileStore.Delete(ctx, atch.ContentPath); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Attachment, error) {
	attachments, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return attachments, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Attachment, error) {
	atch, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Attachment{}, fmt.Errorf("querybyid: %w", err)
	}
	return atch, nil
}

func (b *Business) OpenContent(ctx context.Context, atch Attachment) (io.ReadCloser, error) {
	r, err := b.fileStore.Open(ctx, atch.ContentPath)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			return nil, fmt.Errorf("opencontent: %w", ErrContentNotFound)
		}
		return nil, fmt.Errorf("opencontent: %w", err)
	}
	return r, nil
}
