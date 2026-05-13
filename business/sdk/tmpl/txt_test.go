package tmpl

import (
	"errors"
	"reflect"
	"testing"
)

var txtAllowedRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Target":       {},
}

func TestRequiredVarsTXTAllowsSupportedRoots(t *testing.T) {
	t.Parallel()

	content := []byte(`{{ .Target.Link }} {{ .Employee.FirstName }} {{ .Employee.FirstName }}`)

	vars, err := RequiredVars(content, TXT, txtAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Employee.FirstName", "Target.Link"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsTXTRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	content := []byte(`{{ .Campaign.Label }}`)

	_, err := RequiredVars(content, TXT, txtAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRequiredVarsTXTRejectsUnsupportedSyntax(t *testing.T) {
	t.Parallel()

	content := []byte(`{{ if .Employee.FirstName }}Ivan{{ end }}`)

	_, err := RequiredVars(content, TXT, txtAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRenderTXTRendersPlainText(t *testing.T) {
	t.Parallel()

	content := []byte(`Hello {{ .Employee.FirstName }}`)
	rendered, err := Render(content, TXT, map[string]any{
		"Employee": map[string]any{
			"FirstName": "<Ivan & Co>",
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := "Hello <Ivan & Co>"
	if string(rendered) != expected {
		t.Fatalf("Render = %q, want %q", string(rendered), expected)
	}
}

func TestRenderTXTMapsMissingKey(t *testing.T) {
	t.Parallel()

	content := []byte(`{{ .Employee.FirstName }}`)

	_, err := Render(content, TXT, map[string]any{})
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}
