package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

// Store provides functionality for saving files to the local disk.
type Store struct {
	uploadDir string // Example: "./uploads" or "/tmp/uploads"
	basePath  string // Example: "/uploads" (must be relative for types/url)
}

// New constructs a new local file store.
func New(uploadDir string, basePath string) *Store {
	return &Store{
		uploadDir: uploadDir,
		basePath:  basePath,
	}
}

// Save reads from the provided io.Reader and saves it to the configured
// local directory with a generated UUID name and the provided extension.
// It returns a valid relative url.URL pointing to the saved file.
// If ctx is cancelled during the write, the partially written file is deleted.
func (s *Store) Save(ctx context.Context, r io.Reader, ext string) (url.URL, error) {
	if err := ctx.Err(); err != nil {
		return url.URL{}, fmt.Errorf("save: context already cancelled: %w", err)
	}

	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		return url.URL{}, fmt.Errorf("create upload dir: %w", err)
	}

	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return url.URL{}, fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	// Wrap the reader so each Read call checks ctx first.
	// This lets io.Copy abort if the context is cancelled mid-transfer.
	ctxReader := &contextReader{ctx: ctx, r: r}

	if _, err := io.Copy(out, ctxReader); err != nil {
		// Clean up the partially written file before returning.
		_ = os.Remove(filePath)
		return url.URL{}, fmt.Errorf("copy to file: %w", err)
	}

	// Create a relative web path for the saved file.
	webPath := path.Join(s.basePath, fileName)

	// url.Parse requires a leading slash for relative URLs.
	if webPath != "" && webPath[0] != '/' {
		webPath = "/" + webPath
	}

	return url.Parse(webPath)
}

// Delete removes a file from the local disk given its relative web URL.
func (s *Store) Delete(ctx context.Context, u url.URL) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("delete: context already cancelled: %w", err)
	}

	fileName := path.Base(u.String())
	filePath := filepath.Join(s.uploadDir, fileName)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
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
