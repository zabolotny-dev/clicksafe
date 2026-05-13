package messagebus

import (
	"errors"
	"fmt"
)

var (
	ErrUniqueLabel               = errors.New("Message with this label already exists")
	ErrNotFound                  = errors.New("message not found")
	ErrContentNotFound           = errors.New("message content not found")
	ErrEmptyContent              = errors.New("message content is empty")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
)

type MissingRequiredVarsError struct {
	Vars []string
}

func (e *MissingRequiredVarsError) Error() string {
	return fmt.Sprintf("missing %d required template vars", len(e.Vars))
}
