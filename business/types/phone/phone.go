package phone

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nyaruka/phonenumbers"
)

const phoneRegion = "RU"

type Phone struct {
	value string
}

func (p Phone) String() string {
	return p.value
}

// Equal provides support for the go-cmp package and testing.
func (p Phone) Equal(p2 Phone) bool {
	return p.value == p2.value
}

// MarshalText provides support for logging and any marshal needs.
func (p Phone) MarshalText() ([]byte, error) {
	return []byte(p.value), nil
}

//============================================================================

func Parse(value string) (Phone, error) {
	num, err := phonenumbers.Parse(value, phoneRegion)
	if err != nil {
		return Phone{}, fmt.Errorf("invalid phone %q", err)
	}
	if !phonenumbers.IsValidNumber(num) {
		return Phone{}, fmt.Errorf("invalid phone %q", value)
	}

	return Phone{phonenumbers.Format(num, phonenumbers.E164)}, nil
}

//=============================================================================

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

func ToSQLNullString(n Null) pgtype.Text {
	return pgtype.Text{
		String: n.value,
		Valid:  n.valid,
	}
}

func (n Null) Equal(n2 Null) bool {
	return n.value == n2.value && n.valid == n2.valid
}

func (n Null) MarshalText() ([]byte, error) {
	return []byte(n.value), nil
}

func ParseNull(value string) (Null, error) {
	if value == "" {
		return Null{}, nil
	}

	ph, err := Parse(value)
	if err != nil {
		return Null{}, err
	}

	return Null{ph.String(), true}, nil
}
