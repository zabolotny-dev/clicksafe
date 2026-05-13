package landingbus

import (
	"errors"
)

var (
	ErrUniqueLabel       = errors.New("landing with this label already exists")
	ErrNotFound          = errors.New("landing not found")
	ErrInvalidAttachment = errors.New("invalid attachment")
)
