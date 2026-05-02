package label

import "testing"

func TestParseAllowsRussianEnglishAndDigits(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Отдел 42",
		"Team 7",
		"ООО Ромашка",
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
