package attachmentbus

import (
	"errors"
)

var (
	ErrNotFound                  = errors.New("attachment not found")
	ErrUniqueLabel               = errors.New("attachment with this label already exists")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
	ErrInvalidType               = errors.New("invalid attachment type")
	ErrContentNotFound           = errors.New("content not found")
	ErrEmptyContent              = errors.New("content is empty")
	ErrInUse                     = errors.New("attachment is in use")
)
