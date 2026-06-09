package attachmentbus_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
)

func TestAttachmentType_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    attachmentbus.AttachmentType
		wantErr bool
	}{
		{".html", attachmentbus.Html, false},
		{".png", attachmentbus.Png, false},
		{".pdf", attachmentbus.AttachmentType{}, true},
		{"unknown", attachmentbus.AttachmentType{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := attachmentbus.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAttachmentType_MIMEType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input attachmentbus.AttachmentType
		want  string
	}{
		{attachmentbus.Aac, "audio/aac"},
		{attachmentbus.Docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{attachmentbus.Flac, "audio/flac"},
		{attachmentbus.Gif, "image/gif"},
		{attachmentbus.Html, "text/html"},
		{attachmentbus.Jpeg, "image/jpeg"},
		{attachmentbus.Jpg, "image/jpeg"},
		{attachmentbus.M4a, "audio/mp4"},
		{attachmentbus.Mp3, "audio/mpeg"},
		{attachmentbus.Ogg, "audio/ogg"},
		{attachmentbus.Opus, "audio/ogg; codecs=opus"},
		{attachmentbus.Png, "image/png"},
		{attachmentbus.Pptx, "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{attachmentbus.Txt, "text/plain"},
		{attachmentbus.Wav, "audio/wav"},
		{attachmentbus.Weba, "audio/webm"},
		{attachmentbus.Webm, "audio/webm"},
		{attachmentbus.Webp, "image/webp"},
		{attachmentbus.Xlsx, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{attachmentbus.AttachmentType{}, "application/octet-stream"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input.String(), func(t *testing.T) {
			t.Parallel()
			got := tt.input.MIMEType()
			if got != tt.want {
				t.Errorf("MIMEType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_UnsupportedSyntax(t *testing.T) {
	t.Parallel()

	fs := newFileStorageStub()
	store := newAttachmentStorerStub()
	bus := attachmentbus.NewBusiness(fs, store)

	// Неправильный синтаксис шаблона: незакрытая скобка
	invalidHTML := `<p>Hello {{</p>`
	_, err := bus.UpdateContent(context.Background(), attachmentbus.Attachment{Type: attachmentbus.Html}, attachmentbus.UpdateAttachmentContent{
		Content: strings.NewReader(invalidHTML),
	})
	if err == nil {
		t.Fatal("expected error for invalid template syntax, got nil")
	}
	if !errors.Is(err, attachmentbus.ErrUnsupportedTemplateSyntax) {
		t.Errorf("expected ErrUnsupportedTemplateSyntax, got %v", err)
	}
}
