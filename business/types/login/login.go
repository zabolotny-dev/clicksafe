// Package login represents an administrator login in the system.
package login

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MinLength = 4
	MaxLength = 64
)

// Login represents a validated administrator login.
type Login struct {
	value string
}

// String returns the value of the login.
func (l Login) String() string {
	return l.value
}

// Equal provides support for the go-cmp package and testing.
func (l Login) Equal(l2 Login) bool {
	return l.value == l2.value
}

// MarshalText provides support for logging and any marshal needs.
func (l Login) MarshalText() ([]byte, error) {
	return []byte(l.value), nil
}

// =============================================================================

var loginRegEx = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)

var weakLogins = map[string]struct{}{
	"admin":         {},
	"administrator": {},
	"guest":         {},
	"root":          {},
	"test":          {},
	"user":          {},
}

// Parse parses the string value and returns a login if the value complies
// with the rules for an administrator login.
func Parse(value string) (Login, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return Login{}, fmt.Errorf("login cannot be empty")
	}

	length := utf8.RuneCountInString(value)
	if length < MinLength {
		return Login{}, fmt.Errorf("login too short, must be at least %d characters", MinLength)
	}

	if length > MaxLength {
		return Login{}, fmt.Errorf("login too long, must be %d characters or fewer", MaxLength)
	}

	if !loginRegEx.MatchString(value) {
		return Login{}, fmt.Errorf("invalid login %q", value)
	}

	if _, exists := weakLogins[value]; exists {
		return Login{}, fmt.Errorf("login is too common")
	}

	return Login{value}, nil
}

// MustParse parses the string value and returns a login if the value complies
// with the rules for an administrator login. If an error occurs the function panics.
func MustParse(value string) Login {
	login, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return login
}

// =============================================================================
