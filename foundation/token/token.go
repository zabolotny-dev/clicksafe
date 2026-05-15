package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

var (
	ErrInvalidConfig = errors.New("invalid token config")
	ErrInvalidKey    = errors.New("invalid token hash key")
)

const (
	minTokenBytes     = 32
	minKeyBytes       = 32
	DefaultTokenBytes = 32
)

type Config struct {
	TokenBytes int
	HashKey    string
}

type Manager struct {
	tokenBytes int
	hashKey    []byte
}

func New(cfg Config) (*Manager, error) {
	if cfg.TokenBytes == 0 {
		cfg.TokenBytes = DefaultTokenBytes
	}

	if cfg.TokenBytes < minTokenBytes {
		return nil, ErrInvalidConfig
	}

	if len(cfg.HashKey) < minKeyBytes {
		return nil, ErrInvalidKey
	}

	return &Manager{
		tokenBytes: cfg.TokenBytes,
		hashKey:    []byte(cfg.HashKey),
	}, nil
}

// NewToken returns a random base64url token.
// This is the raw token that goes to the client cookie.
func (m *Manager) NewToken() (string, error) {
	b := make([]byte, m.tokenBytes)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HMACSHA256 returns HMAC-SHA256(token, hashKey) encoded as hex.
// This value is stored in the database.
func (m *Manager) HMACSHA256(value string) string {
	mac := hmac.New(sha256.New, m.hashKey)
	mac.Write([]byte(value))

	return hex.EncodeToString(mac.Sum(nil))
}

// CompareHash checks that raw token corresponds to stored hash.
func (m *Manager) CompareHash(rawToken string, storedHash string) bool {
	actualHash := m.HMACSHA256(rawToken)

	return subtle.ConstantTimeCompare(
		[]byte(actualHash),
		[]byte(storedHash),
	) == 1
}
