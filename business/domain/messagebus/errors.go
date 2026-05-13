package messagebus

import (
	"errors"
)

var (
	ErrUniqueLabel       = errors.New("Message with this label already exists")
	ErrNotFound          = errors.New("message not found")
	ErrInvalidAttachment = errors.New("invalid attachment")
)
