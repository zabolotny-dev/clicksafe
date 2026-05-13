package tmpl

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/xuri/excelize/v2"
)

func extractXLSXVars(content []byte, allowedRoots map[string]struct{}) ([]string, error) {
	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse xlsx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}
	defer file.Close()

	requiredVars := make(map[string]struct{})
	for _, sheet := range file.GetSheetList() {
		rows, ok, err := getXLSXRows(file, sheet)
		if err != nil {
			return nil, fmt.Errorf("read xlsx sheet %q: %w: %v", sheet, ErrUnsupportedTemplateSyntax, err)
		}
		if !ok {
			continue
		}

		for rowIdx, row := range rows {
			for colIdx, cellValue := range row {
				if !containsTemplateAction(cellValue) {
					continue
				}

				cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1, false)
				if err != nil {
					return nil, fmt.Errorf("resolve xlsx cell: %w", err)
				}

				if err := extractXLSXCellVars(sheet, cell, cellValue, allowedRoots, requiredVars); err != nil {
					return nil, err
				}
			}
		}
	}

	vars := make([]string, 0, len(requiredVars))
	for v := range requiredVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)

	return vars, nil
}

func renderXLSX(content []byte, data map[string]any) ([]byte, error) {
	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse xlsx: %w: %v", ErrUnsupportedTemplateSyntax, err)
	}
	defer file.Close()

	for _, sheet := range file.GetSheetList() {
		rows, ok, err := getXLSXRows(file, sheet)
		if err != nil {
			return nil, fmt.Errorf("read xlsx sheet %q: %w: %v", sheet, ErrUnsupportedTemplateSyntax, err)
		}
		if !ok {
			continue
		}

		for rowIdx, row := range rows {
			for colIdx, cellValue := range row {
				if !containsTemplateAction(cellValue) {
					continue
				}

				cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1, false)
				if err != nil {
					return nil, fmt.Errorf("resolve xlsx cell: %w", err)
				}

				rendered, err := renderXLSXCell(sheet, cell, cellValue, data)
				if err != nil {
					return nil, err
				}

				if err := file.SetCellStr(sheet, cell, rendered); err != nil {
					return nil, fmt.Errorf("write xlsx cell %s!%s: %w", sheet, cell, err)
				}
			}
		}
	}

	out, err := file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write xlsx: %w", err)
	}

	return out.Bytes(), nil
}

func getXLSXRows(file *excelize.File, sheet string) ([][]string, bool, error) {
	rows, err := file.GetRows(sheet)
	if err != nil {
		var sheetNotExist excelize.ErrSheetNotExist
		if errors.As(err, &sheetNotExist) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return rows, true, nil
}

func extractXLSXCellVars(sheet string, cell string, value string, allowedRoots map[string]struct{}, requiredVars map[string]struct{}) error {
	tmpl, err := template.New(cell).Parse(value)
	if err != nil {
		return fmt.Errorf("parse xlsx cell %s!%s: %w: %v", sheet, cell, ErrUnsupportedTemplateSyntax, err)
	}

	for _, namedTemplate := range tmpl.Templates() {
		if namedTemplate.Name() != tmpl.Name() {
			return fmt.Errorf("validate xlsx cell %s!%s: %w: named templates are not allowed", sheet, cell, ErrUnsupportedTemplateSyntax)
		}
	}

	if err := validateNode(tmpl.Tree.Root, allowedRoots, requiredVars); err != nil {
		return fmt.Errorf("validate xlsx cell %s!%s: %w: %v", sheet, cell, ErrUnsupportedTemplateSyntax, err)
	}

	return nil
}

func renderXLSXCell(sheet string, cell string, value string, data map[string]any) (string, error) {
	tmpl, err := template.New(cell).Option("missingkey=error").Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse xlsx cell %s!%s: %w: %v", sheet, cell, ErrUnsupportedTemplateSyntax, err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("execute xlsx cell %s!%s: %w: %v", sheet, cell, ErrUnsupportedTemplateSyntax, err)
	}

	return out.String(), nil
}

func containsTemplateAction(value string) bool {
	return strings.Contains(value, "{{")
}
