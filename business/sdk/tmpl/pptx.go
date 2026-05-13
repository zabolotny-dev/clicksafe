package tmpl

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var (
	pptxSplitOpenDelimiter  = regexp.MustCompile(`\{[^{}]*\{`)
	pptxSplitCloseDelimiter = regexp.MustCompile(`\}[^{}]*\}`)
	pptxTemplateAction      = regexp.MustCompile(`(?s)\{\{.*?\}\}`)
	pptxXMLTag              = regexp.MustCompile(`<\s*/?[\w:.-]+(?:\s+[^>]*)?/?>`)
	pptxEntityReplacer      = strings.NewReplacer(
		"&quot;", "\"",
		"&#34;", "\"",
		"&apos;", "'",
		"&#39;", "'",
		"&lt;", "<",
		"&#60;", "<",
		"&gt;", ">",
		"&#62;", ">",
		"&amp;", "&",
		"&#38;", "&",
	)
)

func extractPPTXVars(content []byte, allowedRoots map[string]struct{}) ([]string, error) {
	zipReader, err := openPPTX(content)
	if err != nil {
		return nil, fmt.Errorf("parse pptx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	if err := validatePPTXPackage(zipReader); err != nil {
		return nil, fmt.Errorf("validate pptx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	requiredVars := make(map[string]struct{})
	for _, file := range zipReader.File {
		if !isPPTXTemplatePart(file.Name) {
			continue
		}

		fileContent, err := readPPTXFile(file)
		if err != nil {
			return nil, fmt.Errorf("read pptx part %s: %w", file.Name, err)
		}

		if err := extractPPTXPartVars(file.Name, fileContent, allowedRoots, requiredVars); err != nil {
			return nil, err
		}
	}

	vars := make([]string, 0, len(requiredVars))
	for v := range requiredVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)

	return vars, nil
}

func renderPPTX(content []byte, data map[string]any) ([]byte, error) {
	zipReader, err := openPPTX(content)
	if err != nil {
		return nil, fmt.Errorf("parse pptx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	if err := validatePPTXPackage(zipReader); err != nil {
		return nil, fmt.Errorf("validate pptx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	var out bytes.Buffer
	zipWriter := zip.NewWriter(&out)

	for _, file := range zipReader.File {
		fileContent, err := readPPTXFile(file)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("read pptx part %s: %w", file.Name, err)
		}

		if isPPTXTemplatePart(file.Name) {
			fileContent, err = renderPPTXPart(file.Name, fileContent, data)
			if err != nil {
				zipWriter.Close()
				return nil, err
			}
		}

		if err := writePPTXFile(zipWriter, file, fileContent); err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("write pptx part %s: %w", file.Name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("write pptx: %w", err)
	}

	return out.Bytes(), nil
}

func openPPTX(content []byte) (*zip.Reader, error) {
	return zip.NewReader(bytes.NewReader(content), int64(len(content)))
}

func validatePPTXPackage(zipReader *zip.Reader) error {
	hasContentTypes := false
	hasSlide := false

	for _, file := range zipReader.File {
		switch {
		case file.Name == "[Content_Types].xml":
			hasContentTypes = true
		case isPPTXSlidePart(file.Name):
			hasSlide = true
		}
	}

	if !hasContentTypes {
		return fmt.Errorf("[Content_Types].xml not found")
	}

	if !hasSlide {
		return fmt.Errorf("ppt/slides/*.xml not found")
	}

	return nil
}

func extractPPTXPartVars(name string, content []byte, allowedRoots map[string]struct{}, requiredVars map[string]struct{}) error {
	tmpl, err := template.New(path.Base(name)).Parse(patchPPTXXML(string(content)))
	if err != nil {
		return fmt.Errorf("parse pptx part %s: %w: %v", name, ErrUnsupportedTemplateSyntax, err)
	}

	for _, namedTemplate := range tmpl.Templates() {
		if namedTemplate.Name() != tmpl.Name() {
			return fmt.Errorf("validate pptx part %s: %w: named templates are not allowed", name, ErrUnsupportedTemplateSyntax)
		}
	}

	if err := validateNode(tmpl.Tree.Root, allowedRoots, requiredVars); err != nil {
		return fmt.Errorf("validate pptx part %s: %w: %v", name, ErrUnsupportedTemplateSyntax, err)
	}

	return nil
}

func renderPPTXPart(name string, content []byte, data map[string]any) ([]byte, error) {
	patchedXML := patchPPTXXML(string(content))
	var renderErr error

	renderedXML := pptxTemplateAction.ReplaceAllStringFunc(patchedXML, func(action string) string {
		if renderErr != nil {
			return action
		}

		rendered, err := renderPPTXAction(action, data)
		if err != nil {
			renderErr = fmt.Errorf("render pptx part %s: %w", name, err)
			return action
		}

		return rendered
	})

	if renderErr != nil {
		return nil, renderErr
	}

	return []byte(renderedXML), nil
}

func renderPPTXAction(action string, data map[string]any) (string, error) {
	tmpl, err := template.New("action").Option("missingkey=error").Parse(action)
	if err != nil {
		return "", fmt.Errorf("parse template action: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("execute template action: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, rendered.Bytes()); err != nil {
		return "", fmt.Errorf("escape template action: %w", err)
	}

	return escaped.String(), nil
}

func isPPTXTemplatePart(name string) bool {
	return isPPTXSlidePart(name) || isPPTXNotesPart(name)
}

func isPPTXSlidePart(name string) bool {
	return strings.HasPrefix(name, "ppt/slides/") &&
		strings.HasSuffix(name, ".xml") &&
		!strings.Contains(strings.TrimPrefix(name, "ppt/slides/"), "/")
}

func isPPTXNotesPart(name string) bool {
	return strings.HasPrefix(name, "ppt/notesSlides/") &&
		strings.HasSuffix(name, ".xml") &&
		!strings.Contains(strings.TrimPrefix(name, "ppt/notesSlides/"), "/")
}

func readPPTXFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func writePPTXFile(zipWriter *zip.Writer, file *zip.File, content []byte) error {
	header := file.FileHeader
	writer, err := zipWriter.CreateHeader(&header)
	if err != nil {
		return err
	}

	_, err = writer.Write(content)
	return err
}

func patchPPTXXML(srcXML string) string {
	srcXML = pptxSplitOpenDelimiter.ReplaceAllString(srcXML, "{{")
	srcXML = pptxSplitCloseDelimiter.ReplaceAllString(srcXML, "}}")

	return pptxTemplateAction.ReplaceAllStringFunc(srcXML, func(action string) string {
		action = pptxXMLTag.ReplaceAllString(action, "")
		return pptxEntityReplacer.Replace(action)
	})
}
