package label

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
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

var labelRegEx = regexp.MustCompile(`^[\p{Latin}\p{Cyrillic}][\p{Latin}\p{Cyrillic}0-9' -]{2,254}$`)

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

// Null represents an optional label in the system.
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

	label, err := Parse(value)
	if err != nil {
		return Null{}, err
	}

	return Null{label.String(), true}, nil
}

func (n Null) ToSQLNullString() pgtype.Text {
	return pgtype.Text{
		String: n.value,
		Valid:  n.valid,
	}
}

// =============================================================================
