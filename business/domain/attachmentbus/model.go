package attachmentbus

import (
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Attachment struct {
	ID           uuid.UUID
	Label        label.Label
	Type         AttachmentType
	ContentPath  file.Path
	RequiredVars []string
	Public       bool
	UploadedAt   time.Time
}

type NewAttachment struct {
	Label   label.Label
	Type    AttachmentType
	Content io.Reader
	Public  bool
}

type UpdateAttachment struct {
	Label  *label.Label
	Public *bool
}

type UpdateAttachmentContent struct {
	Content io.Reader
}
