package file

import "testing"

func TestParseAllowsRootRelativePath(t *testing.T) {
	t.Parallel()

	p, err := Parse("/logos/company.png")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := p.String(); got != "/logos/company.png" {
		t.Fatalf("String() = %q, want %q", got, "/logos/company.png")
	}
}

func TestParseRejectsPathWithoutLeadingSlash(t *testing.T) {
	t.Parallel()

	if _, err := Parse("logos/company.png"); err == nil {
		t.Fatal("expected path without leading slash to be rejected")
	}
}

func TestParseRejectsUnnormalizedPath(t *testing.T) {
	t.Parallel()

	if _, err := Parse("/logos/../company.png"); err == nil {
		t.Fatal("expected unnormalized path to be rejected")
	}
}

func TestZeroValuePathIsEmpty(t *testing.T) {
	t.Parallel()

	var p Path
	if !p.IsEmpty() {
		t.Fatal("expected zero value path to be empty")
	}
}

func TestParseNullAllowsEmptyPath(t *testing.T) {
	t.Parallel()

	p, err := ParseNull("   ")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if p.Valid() {
		t.Fatal("expected empty path to be invalid null value")
	}
}

func TestParseNullAllowsValidPath(t *testing.T) {
	t.Parallel()

	p, err := ParseNull("/messages/template.html")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if !p.Valid() {
		t.Fatal("expected path to be valid")
	}

	if got := p.String(); got != "/messages/template.html" {
		t.Fatalf("String() = %q, want %q", got, "/messages/template.html")
	}
}
