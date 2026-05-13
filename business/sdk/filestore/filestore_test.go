package filestore

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

func TestSaveReadDelete(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "/uploads")
	ctx := context.Background()

	p, err := store.Save(ctx, strings.NewReader("hello"), ".txt")
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if !strings.HasPrefix(p.String(), "/uploads/") {
		t.Fatalf("Save returned unexpected path: %q", p.String())
	}

	content, err := store.Read(ctx, p)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if got := string(content); got != "hello" {
		t.Fatalf("Read returned %q, want %q", got, "hello")
	}

	if err := store.Delete(ctx, p); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, err := store.Read(ctx, p); err == nil {
		t.Fatal("expected Read to fail after Delete")
	}
}

func TestOpenReadsSavedFile(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "/uploads")
	ctx := context.Background()

	p, err := store.Save(ctx, strings.NewReader("hello"), ".txt")
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	r, err := store.Open(ctx, p)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	if got := string(content); got != "hello" {
		t.Fatalf("Open returned content %q, want %q", got, "hello")
	}
}

func TestOpenReturnsErrNotFound(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "/uploads")
	ctx := context.Background()

	_, err := store.Open(ctx, file.MustParse("/uploads/missing.txt"))
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Open error = %v, want ErrNotFound", err)
	}
}

func TestOpenRejectsPathOutsidePathPrefix(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "/uploads")
	ctx := context.Background()

	if _, err := store.Open(ctx, file.MustParse("/messages/template.html")); err == nil {
		t.Fatal("expected Open to reject path outside base path")
	}
}

func TestReadRejectsPathOutsidePathPrefix(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "/uploads")
	ctx := context.Background()

	if _, err := store.Read(ctx, file.MustParse("/messages/template.html")); err == nil {
		t.Fatal("expected Read to reject path outside base path")
	}
}

func TestDeleteUsesFullStoredPath(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	store := New(rootDir, "/uploads")
	ctx := context.Background()

	onDiskPath := filepath.Join(rootDir, "nested", "logo.png")
	if err := os.MkdirAll(filepath.Dir(onDiskPath), os.ModePerm); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if err := os.WriteFile(onDiskPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	storedPath := file.MustParse("/uploads/nested/logo.png")
	if err := store.Delete(ctx, storedPath); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, err := os.Stat(onDiskPath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted, stat err = %v", err)
	}
}
