package messagebus

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

type Message struct {
	ID            uuid.UUID
	Label         label.Label
	FromEmail     mail.Address
	FromName      label.Null
	Subject       subject.Null
	HtmlBodyID    uuid.NullUUID
	AttachmentIDs []uuid.UUID
}

type NewMessage struct {
	Label         label.Label
	FromEmail     mail.Address
	FromName      label.Null
	Subject       subject.Null
	HtmlBodyID    uuid.NullUUID
	AttachmentIDs []uuid.UUID
}

type UpdateMessage struct {
	Label         *label.Label
	FromEmail     *mail.Address
	FromName      *label.Null
	Subject       *subject.Null
	HtmlBodyID    *uuid.NullUUID
	AttachmentIDs []uuid.UUID
}
