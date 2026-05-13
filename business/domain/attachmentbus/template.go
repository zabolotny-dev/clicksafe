package attachmentbus

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

func validateAndExtractRequiredVars(content []byte, t AttachmentType) (vars []string, err error) {
	if !t.IsTemplate() {
		return nil, ErrInvalidType
	}

	switch t {
	case Html:
		vars, err = tmpl.RequiredVars(content, tmpl.HTML, allowedTemplateRoots)
	case Docx:
		vars, err = tmpl.RequiredVars(content, tmpl.DOCX, allowedTemplateRoots)
	case Pptx:
		vars, err = tmpl.RequiredVars(content, tmpl.PPTX, allowedTemplateRoots)
	case Xlsx:
		vars, err = tmpl.RequiredVars(content, tmpl.XLSX, allowedTemplateRoots)
	case Txt:
		vars, err = tmpl.RequiredVars(content, tmpl.TXT, allowedTemplateRoots)
	default:
		return nil, ErrInvalidType
	}

	if err != nil {
		if errors.Is(err, tmpl.ErrUnsupportedTemplateSyntax) {
			return nil, fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
		}
		return nil, err
	}

	return vars, nil
}

func render(content []byte, t AttachmentType, data map[string]any) (res []byte, err error) {
	if !t.IsTemplate() {
		return nil, ErrInvalidType
	}

	switch t {
	case Html:
		res, err = tmpl.Render(content, tmpl.HTML, data)
	case Docx:
		res, err = tmpl.Render(content, tmpl.DOCX, data)
	case Pptx:
		res, err = tmpl.Render(content, tmpl.PPTX, data)
	case Xlsx:
		res, err = tmpl.Render(content, tmpl.XLSX, data)
	case Txt:
		res, err = tmpl.Render(content, tmpl.TXT, data)
	default:
		return nil, ErrInvalidType
	}

	if err != nil {
		if errors.Is(err, tmpl.ErrUnsupportedTemplateSyntax) {
			return nil, fmt.Errorf("%w: %v", ErrUnsupportedTemplateSyntax, err)
		}
		return nil, err
	}

	return res, nil
}
