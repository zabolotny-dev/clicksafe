package maxaccountbus

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID
	AdapterID uuid.UUID
	Phone     string
	Label     string
	Status    AccountStatus
	MaxUserID string
	LastError string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginAttempt struct {
	ID        string
	Phone     string
	Label     string
	Status    string
	Error     string
	ExpiresAt string
}

type BeginLogin struct {
	Phone string
	Label string
}

type ConfirmLogin struct {
	AttemptID string
	Code      string
	Password  string
	Label     string
}

type ConfirmPassword struct {
	AttemptID string
	Password  string
	Label     string
}

type UpdateAccount struct {
	Label string
}

type LoginResult struct {
	Attempt LoginAttempt
	Account *Account
}
