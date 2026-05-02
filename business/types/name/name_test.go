package name

import "testing"

func TestParseAllowsRussianAndEnglishPersonNames(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Ян",
		"Иван",
		"Мария-Петрова",
		"O'Connor",
	}

	for _, test := range tests {
		if _, err := Parse(test); err != nil {
			t.Fatalf("Parse(%q) returned error: %v", test, err)
		}
	}
}

func TestParseRejectsDigitsInPersonNames(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Иван2",
		"Al3x",
	}

	for _, test := range tests {
		if _, err := Parse(test); err == nil {
			t.Fatalf("expected Parse(%q) to reject digits", test)
		}
	}
}
