package url

import (
	"fmt"
	"net/url"
)

// URL represents a validated relative web URL (no domain/scheme, just path).
type URL struct {
	value *url.URL
}

// Parse parses a string into a validated relative URL.
func Parse(rawURL string) (URL, error) {
	if rawURL == "" {
		return URL{}, fmt.Errorf("url cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return URL{}, fmt.Errorf("invalid url format: %w", err)
	}

	if u.Scheme != "" || u.Host != "" {
		return URL{}, fmt.Errorf("url must be relative (cannot contain scheme or host)")
	}

	if u.Path != "" && u.Path[0] != '/' {
		return URL{}, fmt.Errorf("relative url must start with '/'")
	}

	return URL{value: u}, nil
}

// String returns the string representation of the URL.
func (u URL) String() string {
	if u.value == nil {
		return ""
	}
	return u.value.String()
}

func (u URL) IsEmpty() bool {
	return u.value == nil
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
