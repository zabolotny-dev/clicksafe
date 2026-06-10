package renderbus_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/usecase/renderbus"
)

// =============================================================================
// Mocks

type fileStorageMock struct {
	content []byte
	readErr error
}

func (m *fileStorageMock) Read(_ context.Context, _ file.Path) ([]byte, error) {
	return m.content, m.readErr
}

type resolverMock struct {
	data       map[string]any
	missing    []string
	resolveErr error
}

func (m *resolverMock) Resolve(_ context.Context, _ uuid.UUID, _ []string) (map[string]any, []string, error) {
	return m.data, m.missing, m.resolveErr
}

// =============================================================================

func Test_Render(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testRender(), "render")
}

// =============================================================================

func testRender() []unittest.Table {
	htmlContent := []byte("<html><body>Hello</body></html>")
	txtContent := []byte("Hello World")

	htmlAatch := attachmentbus.Attachment{
		ID:          uuid.New(),
		Label:       label.MustParse("HTML body"),
		Type:        attachmentbus.Html,
		ContentPath: file.MustParse("/content/file.html"),
	}

	txtAatch := attachmentbus.Attachment{
		ID:          uuid.New(),
		Label:       label.MustParse("TXT body"),
		Type:        attachmentbus.Txt,
		ContentPath: file.MustParse("/content/file.txt"),
	}

	mediaAatch := attachmentbus.Attachment{
		ID:          uuid.New(),
		Label:       label.MustParse("Image"),
		Type:        attachmentbus.Png,
		ContentPath: file.MustParse("/content/file.png"),
	}

	pngContent := []byte{0x89, 0x50, 0x4E, 0x47}

	busOK := renderbus.NewBusiness(
		&fileStorageMock{content: htmlContent},
		&resolverMock{data: map[string]any{}},
	)

	busTxt := renderbus.NewBusiness(
		&fileStorageMock{content: txtContent},
		&resolverMock{data: map[string]any{}},
	)

	busMedia := renderbus.NewBusiness(
		&fileStorageMock{content: pngContent},
		&resolverMock{data: map[string]any{}},
	)

	busNotFound := renderbus.NewBusiness(
		&fileStorageMock{readErr: filestore.ErrNotFound},
		&resolverMock{data: map[string]any{}},
	)

	busReadErr := renderbus.NewBusiness(
		&fileStorageMock{readErr: errors.New("io error")},
		&resolverMock{},
	)

	busMissingVars := renderbus.NewBusiness(
		&fileStorageMock{content: htmlContent},
		&resolverMock{missing: []string{"FirstName"}},
	)

	busResolveErr := renderbus.NewBusiness(
		&fileStorageMock{content: htmlContent},
		&resolverMock{resolveErr: errors.New("resolver error")},
	)

	return []unittest.Table{
		{
			Name:    "renders-html",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busOK.Render(ctx, htmlAatch, uuid.New())
				if err != nil {
					return err
				}
				return bytes.Contains(result, []byte("Hello"))
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "renders-txt",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busTxt.Render(ctx, txtAatch, uuid.New())
				if err != nil {
					return err
				}
				return bytes.Contains(result, []byte("Hello"))
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "passes-through-media",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busMedia.Render(ctx, mediaAatch, uuid.New())
				if err != nil {
					return err
				}
				return bytes.Equal(result, pngContent)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "content-not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNotFound.Render(ctx, htmlAatch, uuid.New())
				return errors.Is(err, renderbus.ErrContentNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "file-read-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busReadErr.Render(ctx, htmlAatch, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "missing-vars-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busMissingVars.Render(ctx, htmlAatch, uuid.New())
				var missingErr *renderbus.MissingRequiredVarsError
				return errors.As(err, &missingErr)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "resolver-error-propagates",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busResolveErr.Render(ctx, htmlAatch, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "invalid-attachment-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Use an unknown type (zero-value AttachmentType is not template/media)
				unknownAatch := attachmentbus.Attachment{
					ID:          uuid.New(),
					Label:       label.MustParse("Unknown type"),
					ContentPath: file.MustParse("/content/unknown.xxx"),
					// Type is zero value — not template, not media
				}
				busInvalid := renderbus.NewBusiness(
					&fileStorageMock{content: []byte("data")},
					&resolverMock{data: map[string]any{}},
				)
				_, err := busInvalid.Render(ctx, unknownAatch, uuid.New())
				return errors.Is(err, renderbus.ErrInvalidType)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "missing-required-vars-error-message",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				e := &renderbus.MissingRequiredVarsError{Vars: []string{"Employee.FirstName"}}
				return len(e.Error()) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
