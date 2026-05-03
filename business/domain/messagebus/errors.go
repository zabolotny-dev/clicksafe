package messagebus

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUniqueLabel               = errors.New("Message with this label already exists")
	ErrNotFound                  = errors.New("message not found")
	ErrContentNotFound           = errors.New("message content not found")
	ErrEmptyContent              = errors.New("message content is empty")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
	ErrResolverNotConfigured     = errors.New("template data resolver is not configured")
)

type MissingRequiredVarsError struct {
	Vars []string
}

func (e *MissingRequiredVarsError) Error() string {
	if e == nil || len(e.Vars) == 0 {
		return "missing required template vars"
	}

	return fmt.Sprintf("missing required template vars: %s", strings.Join(e.Vars, ", "))
}
