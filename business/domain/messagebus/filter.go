package messagebus

import (
	"github.com/google/uuid"
)

type QueryFilter struct {
	ID        *uuid.UUID
	Type      *MessageType
	Label     *string
	FromEmail *string
	FromName  *string
	Subject   *string
}
