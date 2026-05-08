package landingbus

import (
	"errors"
	"fmt"
)

var (
	ErrUniqueLabel               = errors.New("landing with this label already exists")
	ErrNotFound                  = errors.New("landing not found")
	ErrContentNotFound           = errors.New("landing content not found")
	ErrEmptyContent              = errors.New("landing content is empty")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
	ErrResolverNotConfigured     = errors.New("template data resolver is not configured")
)

type MissingRequiredVarsError struct {
	Vars []string
}

func (e *MissingRequiredVarsError) Error() string {
	return fmt.Sprintf("missing %d required template vars", len(e.Vars))
}
