package messagebus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrUniqueLabel       = errors.New("Message with this label already exists")
	ErrNotFound          = errors.New("message not found")
	ErrInvalidAttachment = errors.New("invalid attachment")
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
