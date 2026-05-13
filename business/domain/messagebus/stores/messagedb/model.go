package messagedb

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

func toDBMessage(msg messagebus.Message) sqlc.Message {
	var id *uuid.UUID
	if msg.AttachmentID.Valid {
		id = &msg.AttachmentID.UUID
	}

	return sqlc.Message{
		ID:           msg.ID,
		Label:        msg.Label.String(),
		FromEmail:    msg.FromEmail.String(),
		FromName:     msg.FromName.ToSQLNullString(),
		Subject:      msg.Subject.ToSQLNullString(),
		AttachmentID: id,
	}
}

func toBusMessage(msg sqlc.Message) (messagebus.Message, error) {
	lbl, err := label.Parse(msg.Label)
	if err != nil {
		return messagebus.Message{}, err
	}

	fromEmail, err := mail.ParseAddress(msg.FromEmail)
	if err != nil {
		return messagebus.Message{}, err
	}

	fromName, err := label.ParseNull(msg.FromName.String)
	if err != nil {
		return messagebus.Message{}, err
	}

	sub, err := subject.ParseNull(msg.Subject.String)
	if err != nil {
		return messagebus.Message{}, err
	}

	var id uuid.NullUUID
	if msg.AttachmentID != nil {
		id = uuid.NullUUID{UUID: *msg.AttachmentID, Valid: true}
	}

	return messagebus.Message{
		ID:           msg.ID,
		Label:        lbl,
		FromEmail:    *fromEmail,
		FromName:     fromName,
		Subject:      sub,
		AttachmentID: id,
	}, nil
}

func toBusMessages(messages []sqlc.Message) ([]messagebus.Message, error) {
	busMessages := make([]messagebus.Message, len(messages))

	for i, msg := range messages {
		var err error
		busMessages[i], err = toBusMessage(msg)
		if err != nil {
			return nil, err
		}
	}

	return busMessages, nil
}
