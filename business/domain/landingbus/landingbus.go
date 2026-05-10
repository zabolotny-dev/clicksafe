package landingbus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

type Storer interface {
	Save(ctx context.Context, landing Landing) error
	Update(ctx context.Context, landing Landing) error
	Delete(ctx context.Context, landing Landing) error
	QueryByID(ctx context.Context, id uuid.UUID) (Landing, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Landing, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type FileStorage interface {
	Save(ctx context.Context, r io.Reader, ext string) (file.Path, error)
	Read(ctx context.Context, p file.Path) ([]byte, error)
	Delete(ctx context.Context, p file.Path) error
}

type Resolver interface {
	Resolve(ctx context.Context, targetID uuid.UUID, paths []string) (data map[string]any, missing []string, err error)
}

type Business struct {
	storer    Storer
	fileStore FileStorage
	resolver  Resolver
}

func NewBusiness(storer Storer, fileStore FileStorage, resolver Resolver) *Business {
	return &Business{storer: storer, fileStore: fileStore, resolver: resolver}
}

func (b *Business) Save(ctx context.Context, landing NewLanding) (Landing, error) {
	l := Landing{
		ID:    uuid.New(),
		Label: landing.Label,
	}

	if err := b.storer.Save(ctx, l); err != nil {
		return Landing{}, fmt.Errorf("save: %w", err)
	}

	return l, nil
}

func (b *Business) Update(ctx context.Context, landing Landing, up UpdateLanding) (Landing, error) {
	if up.Label != nil {
		landing.Label = *up.Label
	}

	if err := b.storer.Update(ctx, landing); err != nil {
		return Landing{}, fmt.Errorf("update: %w", err)
	}

	return landing, nil
}

func (b *Business) Delete(ctx context.Context, landing Landing) error {
	if err := b.storer.Delete(ctx, landing); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Landing, error) {
	l, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Landing{}, fmt.Errorf("query: landingID[%s]: %w", id, err)
	}

	return l, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Landing, error) {
	landings, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return landings, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

func (b *Business) SaveContent(ctx context.Context, landing Landing, r io.Reader) (Landing, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Landing{}, fmt.Errorf("savecontent: read: %w", err)
	}

	if len(bytes.TrimSpace(content)) == 0 {
		return Landing{}, ErrEmptyContent
	}

	requiredVars, err := validateAndExtractRequiredVars(content)
	if err != nil {
		return Landing{}, fmt.Errorf("savecontent: %w", err)
	}

	newPath, err := b.fileStore.Save(ctx, bytes.NewReader(content), ".html")
	if err != nil {
		return Landing{}, fmt.Errorf("savecontent: save file: %w", err)
	}

	oldPath := landing.ContentPath

	landing.ContentPath = file.NewNullPath(newPath)
	landing.RequiredVars = requiredVars

	if err := b.storer.Update(ctx, landing); err != nil {
		_ = b.fileStore.Delete(ctx, newPath)
		return Landing{}, fmt.Errorf("savecontent: update: %w", err)
	}

	if oldPath.Valid() && !oldPath.Path().Equal(newPath) {
		_ = b.fileStore.Delete(ctx, oldPath.Path())
	}

	return landing, nil
}

func (b *Business) ReadContent(ctx context.Context, landing Landing) ([]byte, error) {
	if !landing.ContentPath.Valid() {
		return nil, ErrContentNotFound
	}

	content, err := b.fileStore.Read(ctx, landing.ContentPath.Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("readcontent: read file: %w", err)
	}

	return content, nil
}

func (b *Business) Render(ctx context.Context, landing Landing, targetID uuid.UUID) (string, error) {
	if b.resolver == nil {
		return "", fmt.Errorf("render: %w", ErrResolverNotConfigured)
	}

	content, err := b.ReadContent(ctx, landing)
	if err != nil {
		return "", fmt.Errorf("render: read content: %w", err)
	}

	data, missing, err := b.resolver.Resolve(ctx, targetID, landing.RequiredVars)
	if err != nil {
		return "", fmt.Errorf("render: resolve: %w", err)
	}

	if len(missing) > 0 {
		return "", &MissingRequiredVarsError{Vars: append([]string(nil), missing...)}
	}

	tmpl, err := template.New("landing").Option("missingkey=error").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("render: parse template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	if data == nil {
		data = map[string]any{}
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("render: execute template: %w", err)
	}

	return out.String(), nil
}
