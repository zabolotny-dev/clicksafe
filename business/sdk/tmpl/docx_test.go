package tmpl

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

var docxAllowedRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Target":       {},
}

func TestRequiredVarsDOCXAllowsSupportedRoots(t *testing.T) {
	t.Parallel()

	content := testDOCX(t, `{{ .Target.Link }} {{ .Employee.FirstName }} {{ .Employee.FirstName }}`)

	vars, err := RequiredVars(content, DOCX, docxAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Employee.FirstName", "Target.Link"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsDOCXRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	content := testDOCX(t, `{{ .Campaign.Label }}`)

	_, err := RequiredVars(content, DOCX, docxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRenderDOCXReplacesDocumentVariable(t *testing.T) {
	t.Parallel()

	content := testDOCX(t, `Hello {{ .Employee.FirstName }}`)
	rendered, err := Render(content, DOCX, map[string]any{
		"Employee": map[string]any{
			"FirstName": "Ivan",
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	documentXML := readDOCXPart(t, rendered, "word/document.xml")
	if !strings.Contains(documentXML, "Hello Ivan") {
		t.Fatalf("rendered document.xml = %s", documentXML)
	}

	if strings.Contains(documentXML, "{{") {
		t.Fatalf("rendered document.xml still contains template syntax: %s", documentXML)
	}
}

func TestRenderDOCXMapsMissingKey(t *testing.T) {
	t.Parallel()

	content := testDOCX(t, `{{ .Employee.FirstName }}`)

	_, err := Render(content, DOCX, map[string]any{})
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func testDOCX(t *testing.T, textXML string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	writeDOCXPart(t, zipWriter, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)
	writeDOCXPart(t, zipWriter, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`)
	writeDOCXPart(t, zipWriter, "word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`)
	writeDOCXPart(t, zipWriter, "word/document.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>`+textXML+`</w:t></w:r></w:p>
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/>
    </w:sectPr>
  </w:body>
</w:document>`)

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close docx zip: %v", err)
	}

	return buf.Bytes()
}

func writeDOCXPart(t *testing.T, zipWriter *zip.Writer, name string, content string) {
	t.Helper()

	writer, err := zipWriter.Create(name)
	if err != nil {
		t.Fatalf("create docx part %s: %v", name, err)
	}

	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatalf("write docx part %s: %v", name, err)
	}
}

func readDOCXPart(t *testing.T, content []byte, name string) string {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("open docx zip: %v", err)
	}

	for _, file := range zipReader.File {
		if file.Name != name {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open docx part %s: %v", name, err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read docx part %s: %v", name, err)
		}

		return string(data)
	}

	t.Fatalf("docx part %s not found", name)
	return ""
}
