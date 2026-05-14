package renderbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/sdk/tmpl"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

type FileStorage interface {
	Read(ctx context.Context, p file.Path) ([]byte, error)
}

type Resolver interface {
	Resolve(ctx context.Context, targetID uuid.UUID, paths []string) (map[string]any, []string, error)
}

type Business struct {
	fs       FileStorage
	resolver Resolver
}

func NewBusiness(fs FileStorage, resolver Resolver) *Business {
	return &Business{fs: fs, resolver: resolver}
}

func (b *Business) Render(ctx context.Context, atch attachmentbus.Attachment, targetID uuid.UUID) ([]byte, error) {
	content, err := b.fs.Read(ctx, atch.ContentPath)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			return nil, fmt.Errorf("render: %w", ErrContentNotFound)
		}
		return nil, fmt.Errorf("render: %w", err)
	}

	data, missing, err := b.resolver.Resolve(ctx, targetID, atch.RequiredVars)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	if len(missing) > 0 {
		return nil, &MissingRequiredVarsError{Vars: missing}
	}

	res, err := render(content, atch.Type, data)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	return res, nil
}

func render(content []byte, t attachmentbus.AttachmentType, data map[string]any) (res []byte, err error) {
	if t.IsMedia() {
		return content, nil
	}

	switch t {
	case attachmentbus.Html:
		res, err = tmpl.Render(content, tmpl.HTML, data)
	case attachmentbus.Docx:
		res, err = tmpl.Render(content, tmpl.DOCX, data)
	case attachmentbus.Pptx:
		res, err = tmpl.Render(content, tmpl.PPTX, data)
	case attachmentbus.Xlsx:
		res, err = tmpl.Render(content, tmpl.XLSX, data)
	case attachmentbus.Txt:
		res, err = tmpl.Render(content, tmpl.TXT, data)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidType, t)
	}

	if err != nil {
		if errors.Is(err, tmpl.ErrUnsupportedTemplateSyntax) {
			return nil, fmt.Errorf("%w", ErrUnsupportedTemplateSyntax)
		}
		return nil, err
	}

	return res, nil
}
