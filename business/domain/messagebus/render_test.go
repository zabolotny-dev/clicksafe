package messagebus

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func TestRenderUsesResolvedDataDirectly(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	resolver := &resolverStub{
		data: map[string]any{
			"Employee": map[string]any{
				"FirstName": "Ivan",
			},
			"Department": map[string]any{
				"Label": "Human Resources",
			},
		},
	}

	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p><p>{{ .Department.Label }}</p>"),
		},
	}, resolver)

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Department.Label", "Employee.FirstName"}

	rendered, err := business.Render(context.Background(), msg, uuid.New())
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if rendered != "<p>Ivan</p><p>Human Resources</p>" {
		t.Fatalf("Render returned %q", rendered)
	}

	if resolver.calls != 1 {
		t.Fatalf("resolver calls = %d, want 1", resolver.calls)
	}

	if resolver.targetID == uuid.Nil {
		t.Fatal("Render did not forward scope to resolver")
	}

	if len(resolver.paths) != 2 || resolver.paths[0] != "Department.Label" || resolver.paths[1] != "Employee.FirstName" {
		t.Fatalf("Render resolver paths = %v", resolver.paths)
	}
}

func TestRenderReturnsMissingRequiredVars(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p><p>{{ .Department.Label }}</p>"),
		},
	}, &resolverStub{
		missing: []string{"Department.Label"},
	})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName", "Department.Label"}

	_, err := business.Render(context.Background(), msg, uuid.New())
	if err == nil {
		t.Fatal("Render returned nil error")
	}

	var missingErr *MissingRequiredVarsError
	if !errors.As(err, &missingErr) {
		t.Fatalf("Render error = %v, want MissingRequiredVarsError", err)
	}

	if len(missingErr.Vars) != 1 || missingErr.Vars[0] != "Department.Label" {
		t.Fatalf("Render missing vars = %v", missingErr.Vars)
	}
}

func TestRenderReturnsResolverNotConfigured(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p>"),
		},
	}, nil)

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName"}

	_, err := business.Render(context.Background(), msg, uuid.New())
	if !errors.Is(err, ErrResolverNotConfigured) {
		t.Fatalf("Render error = %v, want %v", err, ErrResolverNotConfigured)
	}
}

func TestRenderWrapsResolverError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("resolver error")

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p>"),
		},
	}, &resolverStub{err: testErr})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName"}

	_, err := business.Render(context.Background(), msg, uuid.New())
	if !errors.Is(err, testErr) {
		t.Fatalf("Render error = %v, want %v", err, testErr)
	}
}

func TestRenderReturnsContentNotFound(t *testing.T) {
	t.Parallel()

	business := NewBusiness(nil, &fileStoreStub{}, &resolverStub{})

	_, err := business.Render(context.Background(), Message{}, uuid.New())
	if !errors.Is(err, ErrContentNotFound) {
		t.Fatalf("Render error = %v, want %v", err, ErrContentNotFound)
	}
}

func TestRenderUsesMissingKeySafetyNet(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p>"),
		},
	}, &resolverStub{data: map[string]any{}})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName"}

	_, err := business.Render(context.Background(), msg, uuid.New())
	if err == nil {
		t.Fatal("Render returned nil error")
	}

	if !strings.Contains(err.Error(), `map has no entry for key "Employee"`) {
		t.Fatalf("Render error = %v", err)
	}
}

func TestRenderMapsParseError(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("{{"),
		},
	}, &resolverStub{data: map[string]any{}})

	msg := testMessage(contentPath)

	_, err := business.Render(context.Background(), msg, uuid.New())
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func testMessage(contentPath file.Path) Message {
	return Message{
		ID:          uuid.New(),
		Label:       label.MustParse("Message Render"),
		ContentPath: file.NewNullPath(contentPath),
	}
}

type resolverStub struct {
	data    map[string]any
	missing []string
	err     error
	paths   []string
	targetID uuid.UUID
	calls   int
}

func (s *resolverStub) Resolve(ctx context.Context, targetID uuid.UUID, paths []string) (map[string]any, []string, error) {
	s.calls++
	s.targetID = targetID
	s.paths = append([]string(nil), paths...)

	if s.err != nil {
		return nil, nil, s.err
	}

	return s.data, s.missing, nil
}

type fileStoreStub struct {
	contents map[string][]byte
	readErr  error
}

func (s *fileStoreStub) Save(ctx context.Context, r io.Reader, ext string) (file.Path, error) {
	return file.Path{}, nil
}

func (s *fileStoreStub) Read(ctx context.Context, p file.Path) ([]byte, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}

	content, exists := s.contents[p.String()]
	if !exists {
		return nil, os.ErrNotExist
	}

	return append([]byte(nil), content...), nil
}

func (s *fileStoreStub) Delete(ctx context.Context, p file.Path) error {
	return nil
}
