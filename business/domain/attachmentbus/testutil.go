package attachmentbus

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"

	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

// TestSeedHTMLAttachments creates n HTML attachments using static HTML content.
func TestSeedHTMLAttachments(ctx context.Context, n int, api *Business) ([]Attachment, error) {
	return testSeedAttachments(ctx, n, api, Html, []byte("<html><body>Hello</body></html>"))
}

// TestSeedTxtAttachments creates n TXT attachments using static text content.
func TestSeedTxtAttachments(ctx context.Context, n int, api *Business) ([]Attachment, error) {
	return testSeedAttachments(ctx, n, api, Txt, []byte("Hello World"))
}

// TestSeedImageAttachments creates n PNG attachments using minimal valid PNG bytes.
func TestSeedImageAttachments(ctx context.Context, n int, api *Business) ([]Attachment, error) {
	// Minimal 1x1 transparent PNG
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x62, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}
	return testSeedAttachments(ctx, n, api, Png, png)
}

func testSeedAttachments(ctx context.Context, n int, api *Business, t AttachmentType, content []byte) ([]Attachment, error) {
	atches := make([]Attachment, 0, n)

	idx := rand.Intn(10000)
	for range n {
		idx++

		na := NewAttachment{
			Label:   label.MustParse(fmt.Sprintf("Attachment %d", idx)),
			Type:    t,
			Content: bytes.NewReader(content),
			Public:  false,
		}

		atch, err := api.Save(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding attachment idx[%d]: %w", idx, err)
		}

		atches = append(atches, atch)
	}

	return atches, nil
}
