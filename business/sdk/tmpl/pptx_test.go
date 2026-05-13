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

var pptxAllowedRoots = map[string]struct{}{
	"Organization": {},
	"Department":   {},
	"Employee":     {},
	"Target":       {},
}

func TestRequiredVarsPPTXAllowsSupportedRoots(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`{{ .Target.Link }} {{ .Employee.FirstName }} {{ .Employee.FirstName }}`),
	})

	vars, err := RequiredVars(content, PPTX, pptxAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Employee.FirstName", "Target.Link"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsPPTXExtractsNotes(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml":           testPPTXSlideXML("Hello"),
		"ppt/notesSlides/notesSlide1.xml": testPPTXNotesXML(`{{ .Department.Label }}`),
	})

	vars, err := RequiredVars(content, PPTX, pptxAllowedRoots)
	if err != nil {
		t.Fatalf("RequiredVars returned error: %v", err)
	}

	expected := []string{"Department.Label"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("RequiredVars = %v, want %v", vars, expected)
	}
}

func TestRequiredVarsPPTXRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`{{ .Campaign.Label }}`),
	})

	_, err := RequiredVars(content, PPTX, pptxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRequiredVarsPPTXRejectsUnsupportedSyntax(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`{{ if .Employee.FirstName }}Ivan{{ end }}`),
	})

	_, err := RequiredVars(content, PPTX, pptxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestRenderPPTXReplacesAndEscapesSlideVariable(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`Hello {{ .Employee.FirstName }}`),
	})

	rendered, err := Render(content, PPTX, map[string]any{
		"Employee": map[string]any{
			"FirstName": "<Ivan & Co>",
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	slideXML := readPPTXPart(t, rendered, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "Hello &lt;Ivan &amp; Co&gt;") {
		t.Fatalf("rendered slide XML = %s", slideXML)
	}

	if strings.Contains(slideXML, "{{") {
		t.Fatalf("rendered slide XML still contains template syntax: %s", slideXML)
	}
}

func TestRenderPPTXSupportsSplitRunAction(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`{{ .Employee.</a:t></a:r><a:r><a:t>FirstName }}`),
	})

	rendered, err := Render(content, PPTX, map[string]any{
		"Employee": map[string]any{
			"FirstName": "Ivan",
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	slideXML := readPPTXPart(t, rendered, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "<a:t>Ivan</a:t>") {
		t.Fatalf("rendered slide XML = %s", slideXML)
	}

	if strings.Contains(slideXML, "{{") {
		t.Fatalf("rendered slide XML still contains template syntax: %s", slideXML)
	}
}

func TestRenderPPTXMapsMissingKey(t *testing.T) {
	t.Parallel()

	content := testPPTX(t, map[string]string{
		"ppt/slides/slide1.xml": testPPTXSlideXML(`{{ .Employee.FirstName }}`),
	})

	_, err := Render(content, PPTX, map[string]any{})
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestPPTXMapsInvalidPackage(t *testing.T) {
	t.Parallel()

	content := []byte("not a pptx package")

	if _, err := RequiredVars(content, PPTX, pptxAllowedRoots); !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}

	if _, err := Render(content, PPTX, map[string]any{}); !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("Render error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestPPTXRejectsPackageWithoutSlides(t *testing.T) {
	t.Parallel()

	content := testPPTXZip(t, map[string]string{
		"[Content_Types].xml": testPPTXContentTypesXML(),
	})

	_, err := RequiredVars(content, PPTX, pptxAllowedRoots)
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("RequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func testPPTX(t *testing.T, parts map[string]string) []byte {
	t.Helper()

	allParts := map[string]string{
		"[Content_Types].xml": testPPTXContentTypesXML(),
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`,
		"ppt/presentation.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"></p:presentation>`,
	}

	for name, content := range parts {
		allParts[name] = content
	}

	return testPPTXZip(t, allParts)
}

func testPPTXZip(t *testing.T, parts map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	for name, content := range parts {
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("create pptx part %s: %v", name, err)
		}

		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatalf("write pptx part %s: %v", name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close pptx zip: %v", err)
	}

	return buf.Bytes()
}

func testPPTXContentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
  <Override PartName="/ppt/notesSlides/notesSlide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.notesSlide+xml"/>
</Types>`
}

func testPPTXSlideXML(textXML string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>` + textXML + `</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`
}

func testPPTXNotesXML(textXML string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:notes xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>` + textXML + `</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:notes>`
}

func readPPTXPart(t *testing.T, content []byte, name string) string {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("open pptx zip: %v", err)
	}

	for _, file := range zipReader.File {
		if file.Name != name {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open pptx part %s: %v", name, err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read pptx part %s: %v", name, err)
		}

		return string(data)
	}

	t.Fatalf("pptx part %s not found", name)
	return ""
}
