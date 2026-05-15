package login

import (
	"strings"
	"testing"
)

func TestParseAllowsCommonLogins(t *testing.T) {
	t.Parallel()

	tests := []string{
		"support",
		"john.doe",
		"team-lead_42",
	}

	for _, test := range tests {
		if _, err := Parse(test); err != nil {
			t.Fatalf("Parse(%q) returned error: %v", test, err)
		}
	}
}

func TestParseNormalizesCaseAndSpaces(t *testing.T) {
	t.Parallel()

	login, err := Parse("  Admin.User  ")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := login.String(); got != "admin.user" {
		t.Fatalf("String() = %q, want %q", got, "admin.user")
	}
}

func TestParseRejectsTooShortLogin(t *testing.T) {
	t.Parallel()

	if _, err := Parse("adm"); err == nil {
		t.Fatal("expected too short login to be rejected")
	}
}

func TestParseRejectsCommonLogin(t *testing.T) {
	t.Parallel()

	if _, err := Parse("admin"); err == nil {
		t.Fatal("expected common login to be rejected")
	}
}

func TestParseRejectsDigitFirstCharacter(t *testing.T) {
	t.Parallel()

	if _, err := Parse("1admin"); err == nil {
		t.Fatal("expected digit-first login to be rejected")
	}
}

func TestParseRejectsInvalidCharacters(t *testing.T) {
	t.Parallel()

	if _, err := Parse("admin@mail"); err == nil {
		t.Fatal("expected login with invalid characters to be rejected")
	}
}

func TestParseRejectsTooLongLogin(t *testing.T) {
	t.Parallel()

	if _, err := Parse("a" + strings.Repeat("b", MaxLength)); err == nil {
		t.Fatal("expected too long login to be rejected")
	}
}
