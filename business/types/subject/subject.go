// Package subject represents an email subject in the system.
package subject

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

const MaxLength = 255

// Subject represents a validated email subject.
type Subject struct {
	value string
}

// String returns the value of the subject.
func (s Subject) String() string {
	return s.value
}

// Equal provides support for the go-cmp package and testing.
func (s Subject) Equal(s2 Subject) bool {
	return s.value == s2.value
}

// MarshalText provides support for logging and any marshal needs.
func (s Subject) MarshalText() ([]byte, error) {
	return []byte(s.value), nil
}

// =============================================================================

// Parse parses the string value and returns a subject if the value complies
// with the rules for an email subject.
func Parse(value string) (Subject, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Subject{}, fmt.Errorf("subject cannot be empty")
	}

	if utf8.RuneCountInString(value) > MaxLength {
		return Subject{}, fmt.Errorf("subject too long, must be %d characters or fewer", MaxLength)
	}

	for _, r := range value {
		if unicode.IsControl(r) {
			return Subject{}, fmt.Errorf("subject contains control character")
		}
	}

	return Subject{value}, nil
}

// MustParse parses the string value and returns a subject if the value complies
// with the rules for an email subject. If an error occurs the function panics.
func MustParse(value string) Subject {
	subject, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return subject
}

// =============================================================================

// Null represents an optional email subject.
type Null struct {
	value string
	valid bool
}

func (n Null) String() string {
	if !n.valid {
		return ""
	}

	return n.value
}

func (n Null) Valid() bool {
	return n.valid
}

func (n Null) Equal(n2 Null) bool {
	return n.value == n2.value && n.valid == n2.valid
}

func (n Null) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func ParseNull(value string) (Null, error) {
	if strings.TrimSpace(value) == "" {
		return Null{}, nil
	}

	subject, err := Parse(value)
	if err != nil {
		return Null{}, err
	}

	return Null{subject.String(), true}, nil
}

func (n Null) ToSQLNullString() pgtype.Text {
	return pgtype.Text{
		String: n.value,
		Valid:  n.valid,
	}
}

// =============================================================================
