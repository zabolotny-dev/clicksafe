package tmpl

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	gotemplatedocx "github.com/JJJJJJack/go-template-docx"
)

var (
	docxSplitOpenDelimiter  = regexp.MustCompile(`\{[^{}]*\{`)
	docxSplitCloseDelimiter = regexp.MustCompile(`\}[^{}]*\}`)
	docxTemplateAction      = regexp.MustCompile(`(?s)\{\{.*?\}\}`)
	docxXMLTag              = regexp.MustCompile(`<\s*/?[\w:.-]+(?:\s+[^>]*)?/?>`)
	docxEntityReplacer      = strings.NewReplacer(
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

func extractDOCXVars(content []byte, allowedRoots map[string]struct{}) ([]string, error) {
	docxTemplate, err := gotemplatedocx.NewDocxTemplateFromBytes(content)
	if err != nil {
		return nil, fmt.Errorf("parse docx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	if err := validateDOCXTemplates(content, allowedRoots); err != nil {
		return nil, fmt.Errorf("validate docx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	templateVars, err := docxTemplate.GetTemplateVariables()
	if err != nil {
		return nil, fmt.Errorf("extract docx vars: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	requiredVars := make(map[string]struct{}, len(templateVars))
	for templateVar := range templateVars {
		requiredVar, err := normalizeDOCXVar(templateVar, allowedRoots)
		if err != nil {
			return nil, fmt.Errorf("validate docx var %q: %w: %v", templateVar, ErrUnsupportedTemplateSyntax, err)
		}

		requiredVars[requiredVar] = struct{}{}
	}

	vars := make([]string, 0, len(requiredVars))
	for v := range requiredVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)

	return vars, nil
}

func renderDOCX(content []byte, data map[string]any) ([]byte, error) {
	docxTemplate, err := gotemplatedocx.NewDocxTemplateFromBytes(content)
	if err != nil {
		return nil, fmt.Errorf("parse docx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	if err := docxTemplate.Apply(data); err != nil {
		return nil, fmt.Errorf("execute docx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	return docxTemplate.Bytes(), nil
}

func validateDOCXTemplates(content []byte, allowedRoots map[string]struct{}) error {
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		if !isDOCXTemplatePart(file.Name) {
			continue
		}

		fileContent, err := readDOCXFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file.Name, err)
		}

		tmpl, err := template.New(path.Base(file.Name)).Parse(patchDOCXXML(string(fileContent)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", file.Name, err)
		}

		for _, namedTemplate := range tmpl.Templates() {
			if namedTemplate.Name() != tmpl.Name() {
				return fmt.Errorf("named templates are not allowed")
			}
		}

		if err := validateNode(tmpl.Tree.Root, allowedRoots, make(map[string]struct{})); err != nil {
			return fmt.Errorf("validate %s: %w", file.Name, err)
		}
	}

	return nil
}

func normalizeDOCXVar(templateVar string, allowedRoots map[string]struct{}) (string, error) {
	if !strings.HasPrefix(templateVar, ".") {
		return "", fmt.Errorf("variable references are not allowed")
	}

	segments := strings.Split(strings.TrimPrefix(templateVar, "."), ".")
	if len(segments) < 2 {
		return "", fmt.Errorf("field path must contain at least two segments")
	}

	if _, ok := allowedRoots[segments[0]]; !ok {
		return "", fmt.Errorf("root namespace %q is not allowed", segments[0])
	}

	for _, segment := range segments {
		if segment == "" {
			return "", fmt.Errorf("field path contains an empty segment")
		}
	}

	return strings.Join(segments, "."), nil
}

func isDOCXTemplatePart(name string) bool {
	return strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels")
}

func readDOCXFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func patchDOCXXML(srcXML string) string {
	srcXML = docxSplitOpenDelimiter.ReplaceAllString(srcXML, "{{")
	srcXML = docxSplitCloseDelimiter.ReplaceAllString(srcXML, "}}")

	return docxTemplateAction.ReplaceAllStringFunc(srcXML, func(action string) string {
		action = docxXMLTag.ReplaceAllString(action, "")
		return docxEntityReplacer.Replace(action)
	})
}
