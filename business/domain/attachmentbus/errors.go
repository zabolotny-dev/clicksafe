package attachmentbus

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                  = errors.New("attachment not found")
	ErrUniqueLabel               = errors.New("attachment with this label already exists")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
	ErrInvalidType               = errors.New("invalid attachment type")
	ErrContentNotFound           = errors.New("content not found")
	ErrEmptyContent              = errors.New("content is empty")
)

type MissingRequiredVarsError struct {
	Vars []string
}

func (e *MissingRequiredVarsError) Error() string {
	return fmt.Sprintf("missing %d required template vars", len(e.Vars))
}
