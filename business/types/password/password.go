// Package password represents an administrator password in the system.
package password

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MinLength = 12
	MaxLength = 128
)

var weakPasswords = map[string]struct{}{
	"123456789012":  {},
	"admin123456":   {},
	"admin1234567":  {},
	"administrator": {},
	"changeme123":   {},
	"letmein12345":  {},
	"password123":   {},
	"password1234":  {},
	"password12345": {},
	"qwerty123456":  {},
	"welcome12345":  {},
}

// Password represents a validated administrator password.
type Password struct {
	value string
}

// String returns the value of the password.
func (p Password) String() string {
	return p.value
}

// Equal provides support for the go-cmp package and testing.
func (p Password) Equal(p2 Password) bool {
	return p.value == p2.value
}

// MarshalText provides support for logging and any marshal needs.
func (p Password) MarshalText() ([]byte, error) {
	return []byte(p.value), nil
}

// =============================================================================

// Parse parses the string value and returns a password if the value complies
// with the rules for an administrator password.
func Parse(value string) (Password, error) {
	if value == "" {
		return Password{}, fmt.Errorf("password cannot be empty")
	}

	if strings.TrimSpace(value) != value {
		return Password{}, fmt.Errorf("password must not start or end with spaces")
	}

	length := utf8.RuneCountInString(value)
	if length < MinLength {
		return Password{}, fmt.Errorf("password too short, must be at least %d characters", MinLength)
	}

	if length > MaxLength {
		return Password{}, fmt.Errorf("password too long, must be %d characters or fewer", MaxLength)
	}

	if hasControl(value) {
		return Password{}, fmt.Errorf("password contains control character")
	}

	if repeatedOnly(value) {
		return Password{}, fmt.Errorf("password must contain different characters")
	}

	if characterClasses(value) < 3 {
		return Password{}, fmt.Errorf("password must contain at least three character classes")
	}

	if isCommonWeakPassword(value) {
		return Password{}, fmt.Errorf("password is too common")
	}

	return Password{value}, nil
}

// MustParse parses the string value and returns a password if the value
// complies with the rules for an administrator password. If an error occurs
// the function panics.
func MustParse(value string) Password {
	password, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return password
}

func hasControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}

	return false
}

func repeatedOnly(value string) bool {
	var first rune
	for i, r := range value {
		if i == 0 {
			first = r
			continue
		}

		if r != first {
			return false
		}
	}

	return true
}

func characterClasses(value string) int {
	var hasLower bool
	var hasUpper bool
	var hasDigit bool
	var hasSymbol bool

	for _, r := range value {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	classes := 0
	for _, exists := range []bool{hasLower, hasUpper, hasDigit, hasSymbol} {
		if exists {
			classes++
		}
	}

	return classes
}

func isCommonWeakPassword(value string) bool {
	_, exists := weakPasswords[strings.ToLower(value)]
	return exists
}

// =============================================================================
