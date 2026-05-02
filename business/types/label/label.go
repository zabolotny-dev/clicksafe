// Package label represents a general display label in the system.
package label

import (
	"fmt"
	"regexp"
)

// Label represents a label in the system.
type Label struct {
	value string
}

// String returns the value of the label.
func (l Label) String() string {
	return l.value
}

// Equal provides support for the go-cmp package and testing.
func (l Label) Equal(l2 Label) bool {
	return l.value == l2.value
}

// MarshalText provides support for logging and any marshal needs.
func (l Label) MarshalText() ([]byte, error) {
	return []byte(l.value), nil
}

// =============================================================================

var labelRegEx = regexp.MustCompile(`^[\p{Latin}\p{Cyrillic}][\p{Latin}\p{Cyrillic}0-9' -]{2,19}$`)

// Parse parses the string value and returns a label if the value complies
// with the rules for a label.
func Parse(value string) (Label, error) {
	if !labelRegEx.MatchString(value) {
		return Label{}, fmt.Errorf("invalid label %q", value)
	}

	return Label{value}, nil
}

// MustParse parses the string value and returns a label if the value complies
// with the rules for a label. If an error occurs the function panics.
func MustParse(value string) Label {
	label, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return label
}

// =============================================================================
