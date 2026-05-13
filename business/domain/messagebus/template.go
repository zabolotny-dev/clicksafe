package messagebus

import (
	"errors"
	"fmt"

	"github.com/zabolotny-dev/clicksafe/business/sdk/tmpl"
)

var allowedTemplateRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Target":       {},
}

func validateAndExtractRequiredVars(content []byte) ([]string, error) {
	vars, err := tmpl.RequiredVars(content, tmpl.HTML, allowedTemplateRoots)
	if err != nil {
		if errors.Is(err, tmpl.ErrUnsupportedTemplateSyntax) {
			return nil, fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
		}
		return nil, err
	}

	return vars, nil
}

func renderMessage(content []byte, data map[string]any) ([]byte, error) {
	res, err := tmpl.Render(content, tmpl.HTML, data)
	if err != nil {
		if errors.Is(err, tmpl.ErrUnsupportedTemplateSyntax) {
			return nil, fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
		}
		return nil, err
	}
	return res, nil
}
