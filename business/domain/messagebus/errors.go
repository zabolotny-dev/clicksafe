package messagebus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrUniqueLabel        = errors.New("Message with this label already exists")
	ErrNotFound           = errors.New("message not found")
	ErrInvalidAttachment  = errors.New("invalid attachment")
	ErrMaxHTMLAttachment  = errors.New("max messages do not support html attachments")
	ErrInvalidType        = errors.New("invalid message type")
	ErrFromEmailRequired  = errors.New("message from email is required")
	ErrTextBodyRequired   = errors.New("message text body is required")
	ErrMaxAccountRequired = errors.New("message max account is required")
)

type ErrMissingAttachments struct {
	IDs []uuid.UUID
}

func (e *ErrMissingAttachments) Error() string {
	return fmt.Sprintf("message have missing attachments: %d", len(e.IDs))
}

type ErrDuplicateAttachments struct {
	IDs []uuid.UUID
}

func (e *ErrDuplicateAttachments) Error() string {
	return fmt.Sprintf("message have duplicate attachments: %d", len(e.IDs))
}
