package password

import (
	"strings"
	"testing"
)

func TestParseAllowsStrongPasswords(t *testing.T) {
	t.Parallel()

	tests := []string{
		"StrongPass123",
		"Admin-Panel-42",
		"GoodPass_2026",
	}

	for _, test := range tests {
		if _, err := Parse(test); err != nil {
			t.Fatalf("Parse(%q) returned error: %v", test, err)
		}
	}
}

func TestParseRejectsTooShortPassword(t *testing.T) {
	t.Parallel()

	if _, err := Parse("Aa1!short"); err == nil {
		t.Fatal("expected too short password to be rejected")
	}
}

func TestParseRejectsTooSimplePassword(t *testing.T) {
	t.Parallel()

	if _, err := Parse("onlylowercase"); err == nil {
		t.Fatal("expected one-class password to be rejected")
	}
}

func TestParseRejectsCommonPassword(t *testing.T) {
	t.Parallel()

	if _, err := Parse("Password1234"); err == nil {
		t.Fatal("expected common password to be rejected")
	}
}

func TestParseRejectsRepeatedPassword(t *testing.T) {
	t.Parallel()

	if _, err := Parse("AAAAAAAAAAAA"); err == nil {
		t.Fatal("expected repeated password to be rejected")
	}
}

func TestParseRejectsLeadingOrTrailingSpaces(t *testing.T) {
	t.Parallel()

	if _, err := Parse(" StrongPass123 "); err == nil {
		t.Fatal("expected password with leading or trailing spaces to be rejected")
	}
}

func TestParseRejectsControlCharacters(t *testing.T) {
	t.Parallel()

	if _, err := Parse("StrongPass123\n"); err == nil {
		t.Fatal("expected password with control character to be rejected")
	}
}

func TestParseRejectsTooLongPassword(t *testing.T) {
	t.Parallel()

	if _, err := Parse("Aa1" + strings.Repeat("x", MaxLength)); err == nil {
		t.Fatal("expected too long password to be rejected")
	}
}
