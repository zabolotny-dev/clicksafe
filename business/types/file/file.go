// Package file represents file paths in the system.
package file

import (
	"fmt"
	stdpath "path"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// Path represents a validated root-relative file path.
type Path struct {
	value string
}

// String returns the string value of the path.
func (p Path) String() string {
	return p.value
}

// IsEmpty reports whether the path has no value.
func (p Path) IsEmpty() bool {
	return p.value == ""
}

// Equal provides support for the go-cmp package and testing.
func (p Path) Equal(p2 Path) bool {
	return p.value == p2.value
}

// MarshalText provides support for logging and any marshal needs.
func (p Path) MarshalText() ([]byte, error) {
	return []byte(p.value), nil
}

func (p Path) ToSQLNullString() pgtype.Text {
	return pgtype.Text{
		String: p.String(),
		Valid:  !p.IsEmpty(),
	}
}

// Parse parses the string value and returns a path if the value complies
// with the rules for a root-relative file path.
func Parse(value string) (Path, error) {
	if value == "" {
		return Path{}, fmt.Errorf("path cannot be empty")
	}

	if value[0] != '/' {
		return Path{}, fmt.Errorf("path must start with '/'")
	}

	if strings.Contains(value, "\\") {
		return Path{}, fmt.Errorf("path must use '/' separators")
	}

	cleaned := stdpath.Clean(value)
	if cleaned == "/" {
		return Path{}, fmt.Errorf("path must not be root")
	}

	if cleaned != value {
		return Path{}, fmt.Errorf("path must be normalized")
	}

	return Path{value: value}, nil
}

// MustParse parses the string value and returns a path if the value complies
// with the rules for a root-relative file path. If an error occurs the
// function panics.
func MustParse(value string) Path {
	p, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return p
}

// =============================================================================

// Null represents an optional file path.
type Null struct {
	value Path
	valid bool
}

func (n Null) String() string {
	if !n.valid {
		return ""
	}

	return n.value.String()
}

func (n Null) Valid() bool {
	return n.valid
}

func (n Null) Path() Path {
	return n.value
}

func (n Null) Equal(n2 Null) bool {
	return n.String() == n2.String() && n.valid == n2.valid
}

func (n Null) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func ParseNull(value string) (Null, error) {
	if strings.TrimSpace(value) == "" {
		return Null{}, nil
	}

	p, err := Parse(value)
	if err != nil {
		return Null{}, err
	}

	return Null{value: p, valid: true}, nil
}

func (n Null) ToSQLNullString() pgtype.Text {
	return pgtype.Text{
		String: n.String(),
		Valid:  n.valid,
	}
}
