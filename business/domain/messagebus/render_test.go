package messagebus

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func TestRenderUsesResolvedDataDirectly(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	resolver := &resolverStub{
		result: resolverbus.Result{
			Data: map[string]any{
				"Employee": map[string]any{
					"FirstName": "Ivan",
				},
				"Department": map[string]any{
					"Label": "Human Resources",
				},
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

	rendered, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if rendered != "<p>Ivan</p><p>Human Resources</p>" {
		t.Fatalf("Render returned %q", rendered)
	}

	if resolver.calls != 1 {
		t.Fatalf("resolver calls = %d, want 1", resolver.calls)
	}

	if resolver.scope.TargetID == uuid.Nil {
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
		result: resolverbus.Result{
			Missing: []string{"Department.Label"},
		},
	})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName", "Department.Label"}

	_, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
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

	_, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
	if !errors.Is(err, ErrResolverNotConfigured) {
		t.Fatalf("Render error = %v, want %v", err, ErrResolverNotConfigured)
	}
}

func TestRenderWrapsResolverError(t *testing.T) {
	t.Parallel()

	contentPath := file.MustParse("/uploads/message.html")
	business := NewBusiness(nil, &fileStoreStub{
		contents: map[string][]byte{
			contentPath.String(): []byte("<p>{{ .Employee.FirstName }}</p>"),
		},
	}, &resolverStub{err: resolverbus.ErrEmployeeNotFound})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName"}

	_, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
	if !errors.Is(err, resolverbus.ErrEmployeeNotFound) {
		t.Fatalf("Render error = %v, want %v", err, resolverbus.ErrEmployeeNotFound)
	}
}

func TestRenderReturnsContentNotFound(t *testing.T) {
	t.Parallel()

	business := NewBusiness(nil, &fileStoreStub{}, &resolverStub{})

	_, err := business.Render(context.Background(), Message{}, resolverbus.Scope{TargetID: uuid.New()})
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
	}, &resolverStub{result: resolverbus.Result{Data: map[string]any{}}})

	msg := testMessage(contentPath)
	msg.RequiredVars = []string{"Employee.FirstName"}

	_, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
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
	}, &resolverStub{result: resolverbus.Result{Data: map[string]any{}}})

	msg := testMessage(contentPath)

	_, err := business.Render(context.Background(), msg, resolverbus.Scope{TargetID: uuid.New()})
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
	result resolverbus.Result
	err    error
	paths  []string
	scope  resolverbus.Scope
	calls  int
}

func (s *resolverStub) Resolve(ctx context.Context, scope resolverbus.Scope, paths []string) (resolverbus.Result, error) {
	s.calls++
	s.scope = scope
	s.paths = append([]string(nil), paths...)

	if s.err != nil {
		return resolverbus.Result{}, s.err
	}

	return s.result, nil
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
