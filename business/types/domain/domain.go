package domain

import (
	"fmt"
	"net/url"
	"strings"
)

// Domain represents a validated domain for a phishing page (scheme + host).
type Domain struct {
	value string
}

// String returns the value of the domain.
func (d Domain) String() string {
	return d.value
}

// Equal provides support for the go-cmp package and testing.
func (d Domain) Equal(d2 Domain) bool {
	return d.value == d2.value
}

// MarshalText provides support for logging and any marshal needs.
func (d Domain) MarshalText() ([]byte, error) {
	return []byte(d.value), nil
}

// IsEmpty reports whether d is the empty (unset) domain.
func (d Domain) IsEmpty() bool {
	return d.value == ""
}

// =============================================================================

// Parse parses the string value and returns a domain if the value complies
// with the rules for a phishing domain. The value must contain a valid
// scheme (http or https) and a non-empty host. Trailing slashes are trimmed.
func Parse(value string) (Domain, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Domain{}, fmt.Errorf("domain cannot be empty")
	}

	value = strings.TrimRight(value, "/")

	u, err := url.Parse(value)
	if err != nil {
		return Domain{}, fmt.Errorf("invalid domain %q: %w", value, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return Domain{}, fmt.Errorf("invalid domain scheme %q: must be http or https", u.Scheme)
	}

	if u.Host == "" {
		return Domain{}, fmt.Errorf("invalid domain %q: host is empty", value)
	}

	normalized := u.Scheme + "://" + u.Host

	return Domain{normalized}, nil
}

// MustParse parses the string value and returns a domain if the value complies
// with the rules for a phishing domain. If an error occurs the function panics.
func MustParse(value string) Domain {
	domain, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return domain
}
