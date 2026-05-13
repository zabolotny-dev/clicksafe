package tmpl

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"
)

func extractTXTVars(content []byte, allowedRoots map[string]struct{}) ([]string, error) {
	tmpl, err := template.New("content").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	for _, namedTemplate := range tmpl.Templates() {
		if namedTemplate.Name() != tmpl.Name() {
			return nil, fmt.Errorf("validate template: %w: named templates are not allowed", ErrUnsupportedTemplateSyntax)
		}
	}

	requiredVars := make(map[string]struct{})
	if err := validateNode(tmpl.Tree.Root, allowedRoots, requiredVars); err != nil {
		return nil, fmt.Errorf("validate template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	vars := make([]string, 0, len(requiredVars))
	for v := range requiredVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)

	return vars, nil
}

func renderTXT(content []byte, data map[string]any) ([]byte, error) {
	tmpl, err := template.New("message").Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("execute template: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}

	return out.Bytes(), nil
}
