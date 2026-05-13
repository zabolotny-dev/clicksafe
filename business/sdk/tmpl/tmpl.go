package tmpl

import "errors"

type Format string

var (
	ErrUnsupportedFormat         = errors.New("unsupported format")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
)

const (
	HTML Format = "html"
	DOCX Format = "docx"
	PPTX Format = "pptx"
	XLSX Format = "xlsx"
	TXT  Format = "txt"
	PDF  Format = "pdf"
)

func RequiredVars(content []byte, format Format, allowedRoots map[string]struct{}) ([]string, error) {
	switch format {
	case HTML:
		return extractHTMLVars(content, allowedRoots)
	case DOCX:
		return extractDOCXVars(content, allowedRoots)
	case PPTX:
		return extractPPTXVars(content, allowedRoots)
	case XLSX:
		return extractXLSXVars(content, allowedRoots)
	case TXT:
		return extractTXTVars(content, allowedRoots)
	default:
		return nil, ErrUnsupportedFormat
	}
}

func Render(content []byte, format Format, data map[string]any) ([]byte, error) {
	switch format {
	case HTML:
		return renderHTML(content, data)
	case DOCX:
		return renderDOCX(content, data)
	case PPTX:
		return renderPPTX(content, data)
	case XLSX:
		return renderXLSX(content, data)
	case TXT:
		return renderTXT(content, data)
	default:
		return nil, ErrUnsupportedFormat
	}
}
