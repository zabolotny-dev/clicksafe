package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
	ErrInvalidConfig       = errors.New("invalid argon2 config")
)

// Config holds the parameters for Argon2id hashing.
type Argon2idConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Hasher struct {
	config Argon2idConfig
}

func New(config Argon2idConfig) (*Hasher, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &Hasher{config: config}, nil
}

func validateConfig(c Argon2idConfig) error {
	if c.Memory == 0 ||
		c.Iterations == 0 ||
		c.Parallelism == 0 ||
		c.SaltLength == 0 ||
		c.KeyLength == 0 {
		return ErrInvalidConfig
	}

	if c.SaltLength < 16 {
		return ErrInvalidConfig
	}
	if c.KeyLength < 16 {
		return ErrInvalidConfig
	}

	return nil
}

// Generate creates a salted Argon2id hash of the password using the provided config.
func (h *Hasher) Generate(password string) (string, error) {
	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Iterations,
		h.config.Memory,
		h.config.Parallelism,
		h.config.KeyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.config.Memory,
		h.config.Iterations,
		h.config.Parallelism,
		b64Salt,
		b64Hash,
	), nil
}

// Compare checks if a plaintext password matches the encoded Argon2id hash.
// It parses the parameters from the encoded hash itself.
func (h *Hasher) Compare(password, encodedHash string) (bool, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, ErrInvalidHash
	}

	if vals[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	n, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil || n != 1 {
		return false, ErrInvalidHash
	}

	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	var memory, iterations uint32
	var parallelism uint8

	n, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil || n != 3 {
		return false, ErrInvalidHash
	}

	if memory == 0 || memory > 256*1024 {
		return false, ErrInvalidHash
	}

	if iterations == 0 || iterations > 10 {
		return false, ErrInvalidHash
	}

	if parallelism == 0 || parallelism > 8 {
		return false, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, ErrInvalidHash
	}

	if len(salt) != int(h.config.SaltLength) {
		return false, ErrInvalidHash
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	if len(decodedHash) != int(h.config.KeyLength) {
		return false, ErrInvalidHash
	}

	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(decodedHash)),
	)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
