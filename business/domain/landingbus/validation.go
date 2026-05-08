package landingbus

import (
	"fmt"

	"github.com/zabolotny-dev/clicksafe/business/sdk/htmlval"
)

var allowedTemplateRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Campaign":     {},
	"Target":       {},
}

func validateAndExtractRequiredVars(content []byte) ([]string, error) {
	vars, err := htmlval.ValidateAndExtract(content, allowedTemplateRoots)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	return vars, nil
}
