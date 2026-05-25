package label

import (
	"strings"
	"testing"
)

func TestParseAllowsRussianEnglishAndDigits(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Отдел 42",
		"Team 7",
		"ООО Ромашка",
		"MAX: Смена пароля (со ссылкой)",
		"Отдел ИТ / Безопасность",
		"Администраторы [Группа 1]",
		"Email \"Важное сообщение\"",
	}

	for _, test := range tests {
		if _, err := Parse(test); err != nil {
			t.Fatalf("Parse(%q) returned error: %v", test, err)
		}
	}
}

func TestParseRejectsDigitFirstCharacter(t *testing.T) {
	t.Parallel()

	if _, err := Parse("42 Team"); err == nil {
		t.Fatal("expected digit-first name to be rejected")
	}
}

func TestParseAllowsLongLabel(t *testing.T) {
	t.Parallel()

	value := "Team " + strings.Repeat("A", 250)
	if _, err := Parse(value); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
}

func TestParseRejectsTooLongLabel(t *testing.T) {
	t.Parallel()

	value := "Team " + strings.Repeat("A", 251)
	if _, err := Parse(value); err == nil {
		t.Fatal("expected too long label to be rejected")
	}
}

func TestParseNullAllowsEmptyLabel(t *testing.T) {
	t.Parallel()

	label, err := ParseNull("   ")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if label.Valid() {
		t.Fatal("expected empty label to be invalid null value")
	}
}

func TestParseNullAllowsValidLabel(t *testing.T) {
	t.Parallel()

	label, err := ParseNull("IT Support")
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	if !label.Valid() {
		t.Fatal("expected label to be valid")
	}

	if got := label.String(); got != "IT Support" {
		t.Fatalf("String() = %q, want %q", got, "IT Support")
	}
}
