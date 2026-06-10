package dbtest

import (
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

// BoolPointer returns a *bool from a bool.
func BoolPointer(b bool) *bool {
	return &b
}

// StringPointer returns a *string from a string.
func StringPointer(s string) *string {
	return &s
}

// LabelPointer returns a *label.Label from a string.
func LabelPointer(value string) *label.Label {
	l := label.MustParse(value)
	return &l
}

// NamePointer returns a *name.Name from a string.
func NamePointer(value string) *name.Name {
	n := name.MustParse(value)
	return &n
}

// LoginPointer returns a *login.Login from a string.
func LoginPointer(value string) *login.Login {
	l := login.MustParse(value)
	return &l
}

// PasswordPointer returns a *password.Password from a string.
func PasswordPointer(value string) *password.Password {
	p := password.MustParse(value)
	return &p
}
