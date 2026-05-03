package subject

import (
	"strings"
	"testing"
)

func TestParseAllowsCommonSubjectText(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Security alert: password reset required",
		"Invoice #42 is ready",
		"Q2 report - action required by 18:00",
	}

	for _, test := range tests {
		if _, err := Parse(test); err != nil {
			t.Fatalf("Parse(%q) returned error: %v", test, err)
		}
	}
}

func TestParseTrimsSpaces(t *testing.T) {
	t.Parallel()

	subject, err := Parse("  Hello  ")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := subject.String(); got != "Hello" {
		t.Fatalf("String() = %q, want %q", got, "Hello")
	}
}

func TestParseRejectsEmptySubject(t *testing.T) {
	t.Parallel()

	if _, err := Parse("   "); err == nil {
		t.Fatal("expected empty subject to be rejected")
	}
}

func TestParseRejectsControlCharacters(t *testing.T) {
	t.Parallel()

	if _, err := Parse("Hello\r\nBcc: test@example.com"); err == nil {
		t.Fatal("expected subject with CRLF to be rejected")
	}
}

func TestParseRejectsTooLongSubject(t *testing.T) {
	t.Parallel()

	if _, err := Parse(strings.Repeat("a", MaxLength+1)); err == nil {
		t.Fatal("expected too long subject to be rejected")
	}
}

func TestParseNullAllowsEmptySubject(t *testing.T) {
	t.Parallel()

	subject, err := ParseNull("   ")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if subject.Valid() {
		t.Fatal("expected empty subject to be invalid null value")
	}
}

func TestParseNullAllowsValidSubject(t *testing.T) {
	t.Parallel()

	subject, err := ParseNull("  Hello  ")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if !subject.Valid() {
		t.Fatal("expected subject to be valid")
	}

	if got := subject.String(); got != "Hello" {
		t.Fatalf("String() = %q, want %q", got, "Hello")
	}
}
