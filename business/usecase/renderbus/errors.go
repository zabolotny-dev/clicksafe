package renderbus

import (
	"errors"
	"fmt"
)

var (
	ErrContentNotFound           = errors.New("content not found")
	ErrInvalidType               = errors.New("invalid attachment type")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
)

type MissingRequiredVarsError struct {
	Vars []string
}

func (e *MissingRequiredVarsError) Error() string {
	return fmt.Sprintf("missing required vars: %v", e.Vars)
}
