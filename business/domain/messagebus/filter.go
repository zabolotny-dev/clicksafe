package messagebus

import (
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID        *uuid.UUID
	Label     *string
	FromEmail *string
	FromName  *string
	Subject   *string
}
