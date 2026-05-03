package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

// Store provides functionality for saving files to the local disk.
type Store struct {
	rootDir    string // Example: "./uploads" or "/tmp/uploads"
	pathPrefix string // Example: "/uploads"
}

// New constructs a new local file store.
func New(rootDir string, pathPrefix string) *Store {
	return &Store{
		rootDir:    rootDir,
		pathPrefix: pathPrefix,
	}
}

// Save reads from the provided io.Reader and saves it to the configured
// local directory with a generated UUID name and the provided extension.
// It returns a valid root-relative file path pointing to the saved file.
// If ctx is cancelled during the write, the partially written file is deleted.
func (s *Store) Save(ctx context.Context, r io.Reader, ext string) (file.Path, error) {
	if err := ctx.Err(); err != nil {
		return file.Path{}, fmt.Errorf("save: context already cancelled: %w", err)
	}

	if err := os.MkdirAll(s.rootDir, os.ModePerm); err != nil {
		return file.Path{}, fmt.Errorf("create upload dir: %w", err)
	}

	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.rootDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return file.Path{}, fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Wrap the reader so each Read call checks ctx first.
	// This lets io.Copy abort if the context is cancelled mid-transfer.
	ctxReader := &contextReader{ctx: ctx, r: r}

	if _, err := io.Copy(out, ctxReader); err != nil {
		// Clean up the partially written file before returning.
		_ = os.Remove(filePath)
		return file.Path{}, fmt.Errorf("copy to file: %w", err)
	}

	storedPath := path.Join(s.normalizedPathPrefix(), fileName)

	return file.Parse(storedPath)
}

// Read returns file contents for the provided stored path.
func (s *Store) Read(ctx context.Context, p file.Path) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read: context already cancelled: %w", err)
	}

	filePath, err := s.resolveDiskPath(p)
	if err != nil {
		return nil, fmt.Errorf("read: resolve path: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return content, nil
}

// Delete removes a file from the local disk given its stored path.
func (s *Store) Delete(ctx context.Context, p file.Path) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("delete: context already cancelled: %w", err)
	}

	filePath, err := s.resolveDiskPath(p)
	if err != nil {
		return fmt.Errorf("delete: resolve path: %w", err)
	}

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func (s *Store) normalizedPathPrefix() string {
	pathPrefix := strings.TrimSpace(s.pathPrefix)
	if pathPrefix == "" {
		return "/"
	}

	if pathPrefix[0] != '/' {
		pathPrefix = "/" + pathPrefix
	}

	return path.Clean(pathPrefix)
}

func (s *Store) resolveDiskPath(p file.Path) (string, error) {
	pathPrefix := s.normalizedPathPrefix()
	storedPath := p.String()

	if pathPrefix != "/" && storedPath != pathPrefix && !strings.HasPrefix(storedPath, pathPrefix+"/") {
		return "", fmt.Errorf("path %q is outside path prefix %q", storedPath, pathPrefix)
	}

	relPath := strings.TrimPrefix(storedPath, pathPrefix)
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath == "" {
		return "", fmt.Errorf("path must point to a file")
	}

	return filepath.Join(s.rootDir, filepath.FromSlash(relPath)), nil
}

// contextReader wraps an io.Reader and checks ctx.Err() before every read.
// This allows io.Copy to abort as soon as the context is cancelled.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}
