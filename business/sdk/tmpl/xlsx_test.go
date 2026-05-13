package tmpl

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

var xlsxAllowedRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Target":       {},
}

func TestRequiredVarsXLSXAllowsSupportedRoots(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "{{ .Target.Link }}",
			"B1": "{{ .Employee.FirstName }} {{ .Employee.FirstName }}",
		},
	})

	vars, err := RequiredVars(content, XLSX, xlsxAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Employee.FirstName", "Target.Link"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsXLSXExtractsMultipleSheets(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "{{ .Employee.FirstName }}",
		},
		"Details": {
			"A1": "{{ .Department.Label }}",
		},
	})

	vars, err := RequiredVars(content, XLSX, xlsxAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Department.Label", "Employee.FirstName"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsXLSXRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "{{ .Campaign.Label }}",
		},
	})

	_, err := RequiredVars(content, XLSX, xlsxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRequiredVarsXLSXRejectsUnsupportedSyntax(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "{{ if .Employee.FirstName }}Ivan{{ end }}",
		},
	})

	_, err := RequiredVars(content, XLSX, xlsxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRenderXLSXReplacesCellVariables(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "Hello {{ .Employee.FirstName }}",
			"B1": "Static",
			"C1": "{{ .Target.Link }}",
		},
	})

	rendered, err := Render(content, XLSX, map[string]any{
		"Employee": map[string]any{
			"FirstName": "Ivan",
		},
		"Target": map[string]any{
			"Link": "https://example.test",
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if value := readXLSXCell(t, rendered, "Sheet1", "A1"); value != "Hello Ivan" {
		t.Fatalf("rendered Sheet1!A1 = %q, want %q", value, "Hello Ivan")
	}

	if value := readXLSXCell(t, rendered, "Sheet1", "B1"); value != "Static" {
		t.Fatalf("rendered Sheet1!B1 = %q, want %q", value, "Static")
	}

	if value := readXLSXCell(t, rendered, "Sheet1", "C1"); value != "https://example.test" {
		t.Fatalf("rendered Sheet1!C1 = %q, want %q", value, "https://example.test")
	}
}

func TestRenderXLSXMapsMissingKey(t *testing.T) {
	t.Parallel()

	content := testXLSX(t, map[string]map[string]string{
		"Sheet1": {
			"A1": "{{ .Employee.FirstName }}",
		},
	})

	_, err := Render(content, XLSX, map[string]any{})
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestXLSXMapsInvalidWorkbook(t *testing.T) {
	t.Parallel()

	content := []byte("not an xlsx workbook")

	if _, err := RequiredVars(content, XLSX, xlsxAllowedRoots); !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}

	if _, err := Render(content, XLSX, map[string]any{}); !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func testXLSX(t *testing.T, sheets map[string]map[string]string) []byte {
	t.Helper()

	file := excelize.NewFile()
	defer file.Close()

	for sheet, cells := range sheets {
		if sheet != "Sheet1" {
			if _, err := file.NewSheet(sheet); err != nil {
				t.Fatalf("create xlsx sheet %s: %v", sheet, err)
			}
		}

		for cell, value := range cells {
			if err := file.SetCellStr(sheet, cell, value); err != nil {
				t.Fatalf("write xlsx cell %s!%s: %v", sheet, cell, err)
			}
		}
	}

	buf, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("write xlsx workbook: %v", err)
	}

	return buf.Bytes()
}

func readXLSXCell(t *testing.T, content []byte, sheet string, cell string) string {
	t.Helper()

	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("open xlsx workbook: %v", err)
	}
	defer file.Close()

	value, err := file.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("read xlsx cell %s!%s: %v", sheet, cell, err)
	}

	return value
}
