package messagebus

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

type Message struct {
	ID            uuid.UUID
	Type          MessageType
	Label         label.Label
	FromEmail     mail.Address
	FromName      label.Null
	Subject       subject.Null
	HtmlBodyID    uuid.NullUUID
	TextBodyID    uuid.NullUUID
	MaxAccountID  uuid.NullUUID
	AttachmentIDs []uuid.UUID
}

type NewMessage struct {
	Type          MessageType
	Label         label.Label
	FromEmail     mail.Address
	FromName      label.Null
	Subject       subject.Null
	HtmlBodyID    uuid.NullUUID
	TextBodyID    uuid.NullUUID
	MaxAccountID  uuid.NullUUID
	AttachmentIDs []uuid.UUID
}

type UpdateMessage struct {
	Type          *MessageType
	Label         *label.Label
	FromEmail     *mail.Address
	FromName      *label.Null
	Subject       *subject.Null
	HtmlBodyID    *uuid.NullUUID
	TextBodyID    *uuid.NullUUID
	MaxAccountID  *uuid.NullUUID
	AttachmentIDs []uuid.UUID
}
