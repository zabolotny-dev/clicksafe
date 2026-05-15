package adminbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

type Admin struct {
	ID           uuid.UUID
	FirstName    name.Name
	LastName     name.Name
	Login        login.Login
	PasswordHash string
	CreatedAt    time.Time
}

type NewAdmin struct {
	FirstName name.Name
	LastName  name.Name
	Login     login.Login
	Password  password.Password
}

type UpdateAdmin struct {
	FirstName *name.Name
	LastName  *name.Name
	Login     *login.Login
	Password  *password.Password
}
