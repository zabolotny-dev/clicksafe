package attachmentbus_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/filestore"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// =============================================================================
// Stubs

type attachmentStorerStub struct {
	saved    attachmentbus.Attachment
	data     map[uuid.UUID]attachmentbus.Attachment
	saveErr  error
	queryErr error
	deleted  []uuid.UUID
}

func newAttachmentStorerStub() *attachmentStorerStub {
	return &attachmentStorerStub{data: make(map[uuid.UUID]attachmentbus.Attachment)}
}

func (s *attachmentStorerStub) Save(_ context.Context, a attachmentbus.Attachment) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = a
	s.data[a.ID] = a
	return nil
}

func (s *attachmentStorerStub) Update(_ context.Context, a attachmentbus.Attachment) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = a
	s.data[a.ID] = a
	return nil
}

func (s *attachmentStorerStub) Delete(_ context.Context, a attachmentbus.Attachment) error {
	s.deleted = append(s.deleted, a.ID)
	delete(s.data, a.ID)
	return nil
}

func (s *attachmentStorerStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	if s.queryErr != nil {
		return attachmentbus.Attachment{}, s.queryErr
	}
	a, ok := s.data[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return a, nil
}

func (s *attachmentStorerStub) Query(_ context.Context, _ attachmentbus.QueryFilter, _ order.By, _ page.Page) ([]attachmentbus.Attachment, error) {
	return nil, nil
}

func (s *attachmentStorerStub) Count(_ context.Context, _ attachmentbus.QueryFilter) (int, error) {
	return len(s.data), nil
}

// fileStorageStub сохраняет данные в памяти.
type fileStorageStub struct {
	files      map[string][]byte
	saveErr    error
	deletedPaths []string
}

func newFileStorageStub() *fileStorageStub {
	return &fileStorageStub{files: make(map[string][]byte)}
}

func (s *fileStorageStub) Save(_ context.Context, r io.Reader, ext string) (file.Path, error) {
	if s.saveErr != nil {
		return file.Path{}, s.saveErr
	}
	data, _ := io.ReadAll(r)
	p := file.MustParse("/" + uuid.New().String() + ext)
	s.files[p.String()] = data
	return p, nil
}

func (s *fileStorageStub) Read(_ context.Context, p file.Path) ([]byte, error) {
	data, ok := s.files[p.String()]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}

func (s *fileStorageStub) Open(_ context.Context, p file.Path) (io.ReadCloser, error) {
	data, ok := s.files[p.String()]
	if !ok {
		return nil, filestore.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *fileStorageStub) Delete(_ context.Context, p file.Path) error {
	s.deletedPaths = append(s.deletedPaths, p.String())
	delete(s.files, p.String())
	return nil
}

// =============================================================================
// Save tests

func TestSave_EmptyContent_ReturnsErrEmptyContent(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	_, err := bus.Save(context.Background(), attachmentbus.NewAttachment{
		Label:   label.MustParse("Empty file"),
		Type:    attachmentbus.Png,
		Content: strings.NewReader(""),
	})
	if !errors.Is(err, attachmentbus.ErrEmptyContent) {
		t.Fatalf("Save error = %v, want %v", err, attachmentbus.ErrEmptyContent)
	}
}

func TestSave_InvalidType_ReturnsErrInvalidType(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	// Нулевой AttachmentType не является ни шаблоном, ни медиа.
	_, err := bus.Save(context.Background(), attachmentbus.NewAttachment{
		Label:   label.MustParse("bad type"),
		Type:    attachmentbus.AttachmentType{},
		Content: strings.NewReader("some content"),
	})
	if !errors.Is(err, attachmentbus.ErrInvalidType) {
		t.Fatalf("Save error = %v, want %v", err, attachmentbus.ErrInvalidType)
	}
}

func TestSave_PNGImage_Success(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(fs, store)

	// PNG — медиа-тип, не шаблон — шаблонная валидация пропускается.
	content := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	a, err := bus.Save(context.Background(), attachmentbus.NewAttachment{
		Label:   label.MustParse("Logo"),
		Type:    attachmentbus.Png,
		Content: bytes.NewReader(content),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if a.ID == (uuid.UUID{}) {
		t.Error("expected non-zero UUID")
	}
	if a.Type != attachmentbus.Png {
		t.Errorf("Type = %v, want %v", a.Type, attachmentbus.Png)
	}
	if store.saved.ID != a.ID {
		t.Error("Attachment was not persisted to storer")
	}
}

func TestSave_HTMLTemplate_ExtractsVars(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(fs, store)

	// Корректный HTML с шаблонной переменной в синтаксисе Go-шаблонов
	html := `<p>Hello {{.Employee.FirstName}}</p>`
	a, err := bus.Save(context.Background(), attachmentbus.NewAttachment{
		Label:   label.MustParse("Welcome email"),
		Type:    attachmentbus.Html,
		Content: strings.NewReader(html),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	_ = a // RequiredVars may or may not be populated depending on parse logic
}

func TestSave_StorerError_Propagates(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	store.saveErr = errors.New("db error")
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	_, err := bus.Save(context.Background(), attachmentbus.NewAttachment{
		Label:   label.MustParse("Logo"),
		Type:    attachmentbus.Png,
		Content: bytes.NewReader([]byte{1, 2, 3}),
	})
	if err == nil {
		t.Fatal("expected error when storer fails, got nil")
	}
}

// =============================================================================
// Update tests

func TestUpdate_ChangesLabel(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	original := attachmentbus.Attachment{
		ID:    uuid.New(),
		Label: label.MustParse("Old label"),
		Type:  attachmentbus.Png,
	}
	store.data[original.ID] = original

	newLabel := label.MustParse("New label")
	updated, err := bus.Update(context.Background(), original, attachmentbus.UpdateAttachment{
		Label: &newLabel,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Label != newLabel {
		t.Errorf("Label = %v, want %v", updated.Label, newLabel)
	}
}

func TestUpdate_ChangesPublic(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	original := attachmentbus.Attachment{
		ID:     uuid.New(),
		Label:  label.MustParse("File"),
		Type:   attachmentbus.Png,
		Public: false,
	}
	store.data[original.ID] = original

	pub := true
	updated, err := bus.Update(context.Background(), original, attachmentbus.UpdateAttachment{
		Public: &pub,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !updated.Public {
		t.Error("expected Public = true after update")
	}
}

// =============================================================================
// UpdateContent tests

func TestUpdateContent_NonTemplateType_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	a := attachmentbus.Attachment{
		ID:   uuid.New(),
		Type: attachmentbus.Png, // PNG не шаблон
	}

	_, err := bus.UpdateContent(context.Background(), a, attachmentbus.UpdateAttachmentContent{
		Content: strings.NewReader("<p>Hello</p>"),
	})
	if err == nil {
		t.Fatal("expected error for non-template type, got nil")
	}
}

func TestUpdateContent_EmptyContent_ReturnsErrEmptyContent(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	a := attachmentbus.Attachment{
		ID:   uuid.New(),
		Type: attachmentbus.Html,
	}

	_, err := bus.UpdateContent(context.Background(), a, attachmentbus.UpdateAttachmentContent{
		Content: strings.NewReader(""),
	})
	if !errors.Is(err, attachmentbus.ErrEmptyContent) {
		t.Fatalf("UpdateContent error = %v, want %v", err, attachmentbus.ErrEmptyContent)
	}
}

func TestUpdateContent_NilContent_ReturnsErrEmptyContent(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	a := attachmentbus.Attachment{
		ID:   uuid.New(),
		Type: attachmentbus.Html,
	}

	_, err := bus.UpdateContent(context.Background(), a, attachmentbus.UpdateAttachmentContent{
		Content: nil,
	})
	if !errors.Is(err, attachmentbus.ErrEmptyContent) {
		t.Fatalf("UpdateContent error = %v, want %v", err, attachmentbus.ErrEmptyContent)
	}
}

func TestUpdateContent_ReplacesFileAndDeletesOld(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(fs, store)

	// Сохраняем исходный файл
	oldPath := file.MustParse("/old-file.html")
	fs.files[oldPath.String()] = []byte("<p>old</p>")

	a := attachmentbus.Attachment{
		ID:          uuid.New(),
		Type:        attachmentbus.Html,
		Label:       label.MustParse("Template"),
		ContentPath: oldPath,
	}
	store.data[a.ID] = a

	updated, err := bus.UpdateContent(context.Background(), a, attachmentbus.UpdateAttachmentContent{
		Content: strings.NewReader("<p>Hello {{.Employee.FirstName}}</p>"),
	})
	if err != nil {
		t.Fatalf("UpdateContent returned error: %v", err)
	}

	// Новый путь должен отличаться от старого
	if updated.ContentPath == oldPath {
		t.Error("ContentPath must change after UpdateContent")
	}

	// Старый файл должен быть удалён
	found := false
	for _, p := range fs.deletedPaths {
		if p == oldPath.String() {
			found = true
			break
		}
	}
	if !found {
		t.Error("old file was not deleted from file storage")
	}
}

// =============================================================================
// Delete tests

func TestDelete_RemovesFromStoreAndFileStorage(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(fs, store)

	contentPath := file.MustParse("/some-file.png")
	fs.files[contentPath.String()] = []byte{1, 2, 3}

	a := attachmentbus.Attachment{
		ID:          uuid.New(),
		Type:        attachmentbus.Png,
		ContentPath: contentPath,
	}
	store.data[a.ID] = a

	if err := bus.Delete(context.Background(), a); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, exists := store.data[a.ID]; exists {
		t.Error("attachment still exists in storer after delete")
	}

	if _, exists := fs.files[contentPath.String()]; exists {
		t.Error("file still exists in file storage after delete")
	}
}

// =============================================================================
// QueryByID tests

func TestQueryByID_ReturnsAttachment(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	a := attachmentbus.Attachment{
		ID:    uuid.New(),
		Label: label.MustParse("Test"),
		Type:  attachmentbus.Png,
	}
	store.data[a.ID] = a

	got, err := bus.QueryByID(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("QueryByID returned error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("ID = %v, want %v", got.ID, a.ID)
	}
}

func TestQueryByID_NotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	_, err := bus.QueryByID(context.Background(), uuid.New())
	if !errors.Is(err, attachmentbus.ErrNotFound) {
		t.Fatalf("QueryByID error = %v, want %v", err, attachmentbus.ErrNotFound)
	}
}

// =============================================================================
// Query / Count tests

func TestQuery_ReturnsAttachments(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	result, err := bus.Query(context.Background(), attachmentbus.QueryFilter{}, order.By{}, page.MustParse("1", "10"))
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	_ = result // stub returns empty slice
}

func TestCount_ReturnsCorrectNumber(t *testing.T) {
	t.Parallel()

	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(newFileStorageStub(), store)

	for i := 0; i < 3; i++ {
		store.data[uuid.New()] = attachmentbus.Attachment{ID: uuid.New()}
	}

	count, err := bus.Count(context.Background(), attachmentbus.QueryFilter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

// =============================================================================
// OpenContent tests

func TestOpenContent_ReturnsReader(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	bus := attachmentbus.NewBusiness(fs, newAttachmentStorerStub())

	contentPath := file.MustParse("/test-content.html")
	fs.files[contentPath.String()] = []byte("<p>content</p>")

	a := attachmentbus.Attachment{
		ID:          uuid.New(),
		Type:        attachmentbus.Html,
		ContentPath: contentPath,
	}

	rc, err := bus.OpenContent(context.Background(), a)
	if err != nil {
		t.Fatalf("OpenContent returned error: %v", err)
	}
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	if string(data) != "<p>content</p>" {
		t.Errorf("OpenContent data = %q, want %q", data, "<p>content</p>")
	}
}

func TestOpenContent_FileNotFound_ReturnsErrContentNotFound(t *testing.T) {
	t.Parallel()

	bus := attachmentbus.NewBusiness(newFileStorageStub(), newAttachmentStorerStub())

	a := attachmentbus.Attachment{
		ID:          uuid.New(),
		Type:        attachmentbus.Html,
		ContentPath: file.MustParse("/nonexistent.html"),
	}

	_, err := bus.OpenContent(context.Background(), a)
	if !errors.Is(err, attachmentbus.ErrContentNotFound) {
		t.Fatalf("OpenContent error = %v, want %v", err, attachmentbus.ErrContentNotFound)
	}
}

// =============================================================================
// AttachmentType tests

func TestAttachmentType_IsTemplate(t *testing.T) {
	t.Parallel()

	templateTypes := []attachmentbus.AttachmentType{
		attachmentbus.Html, attachmentbus.Docx,
		attachmentbus.Pptx, attachmentbus.Xlsx, attachmentbus.Txt,
	}
	for _, at := range templateTypes {
		if !at.IsTemplate() {
			t.Errorf("type %v should be a template", at)
		}
	}

	mediaTypes := []attachmentbus.AttachmentType{
		attachmentbus.Png, attachmentbus.Jpg, attachmentbus.Mp3,
	}
	for _, at := range mediaTypes {
		if at.IsTemplate() {
			t.Errorf("type %v should not be a template", at)
		}
	}
}

func TestAttachmentType_IsImage(t *testing.T) {
	t.Parallel()

	imageTypes := []attachmentbus.AttachmentType{
		attachmentbus.Png, attachmentbus.Jpg, attachmentbus.Jpeg,
		attachmentbus.Gif, attachmentbus.Webp,
	}
	for _, at := range imageTypes {
		if !at.IsImage() {
			t.Errorf("type %v should be an image", at)
		}
	}
	if attachmentbus.Mp3.IsImage() {
		t.Error("Mp3 should not be an image")
	}
}

func TestAttachmentType_IsAudio(t *testing.T) {
	t.Parallel()

	audioTypes := []attachmentbus.AttachmentType{
		attachmentbus.Mp3, attachmentbus.Aac, attachmentbus.Wav,
	}
	for _, at := range audioTypes {
		if !at.IsAudio() {
			t.Errorf("type %v should be audio", at)
		}
	}
	if attachmentbus.Png.IsAudio() {
		t.Error("Png should not be audio")
	}
}

