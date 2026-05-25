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
	var htmlBodyID *uuid.UUID
	if msg.HtmlBodyID.Valid {
		htmlBodyID = &msg.HtmlBodyID.UUID
	}

	var textBodyID *uuid.UUID
	if msg.TextBodyID.Valid {
		textBodyID = &msg.TextBodyID.UUID
	}

	var maxAccountID *uuid.UUID
	if msg.MaxAccountID.Valid {
		maxAccountID = &msg.MaxAccountID.UUID
	}

	return sqlc.Message{
		ID:           msg.ID,
		Type:         msg.Type.String(),
		Label:        msg.Label.String(),
		FromEmail:    msg.FromEmail.Address,
		FromName:     msg.FromName.ToSQLNullString(),
		Subject:      msg.Subject.ToSQLNullString(),
		HtmlBodyID:   htmlBodyID,
		TextBodyID:   textBodyID,
		MaxAccountID: maxAccountID,
	}
}

func toBusMessage(msg sqlc.Message, attachmentIDs []uuid.UUID) (messagebus.Message, error) {
	lbl, err := label.Parse(msg.Label)
	if err != nil {
		return messagebus.Message{}, err
	}

	msgType, err := messagebus.ParseMessageType(msg.Type)
	if err != nil {
		return messagebus.Message{}, err
	}

	var fromEmail mail.Address
	if msg.FromEmail != "" {
		parsed, err := mail.ParseAddress(msg.FromEmail)
		if err != nil {
			return messagebus.Message{}, err
		}
		fromEmail = *parsed
	}

	fromName, err := label.ParseNull(msg.FromName.String)
	if err != nil {
		return messagebus.Message{}, err
	}

	sub, err := subject.ParseNull(msg.Subject.String)
	if err != nil {
		return messagebus.Message{}, err
	}

	var htmlBodyID uuid.NullUUID
	if msg.HtmlBodyID != nil {
		htmlBodyID = uuid.NullUUID{UUID: *msg.HtmlBodyID, Valid: true}
	}

	var textBodyID uuid.NullUUID
	if msg.TextBodyID != nil {
		textBodyID = uuid.NullUUID{UUID: *msg.TextBodyID, Valid: true}
	}

	var maxAccountID uuid.NullUUID
	if msg.MaxAccountID != nil {
		maxAccountID = uuid.NullUUID{UUID: *msg.MaxAccountID, Valid: true}
	}

	return messagebus.Message{
		ID:            msg.ID,
		Type:          msgType,
		Label:         lbl,
		FromEmail:     fromEmail,
		FromName:      fromName,
		Subject:       sub,
		HtmlBodyID:    htmlBodyID,
		TextBodyID:    textBodyID,
		MaxAccountID:  maxAccountID,
		AttachmentIDs: attachmentIDs,
	}, nil
}

func toBusMessages(messages []sqlc.Message, attachmentsByMsg map[uuid.UUID][]uuid.UUID) ([]messagebus.Message, error) {
	busMessages := make([]messagebus.Message, len(messages))

	for i, msg := range messages {
		var err error
		busMessages[i], err = toBusMessage(msg, attachmentsByMsg[msg.ID])
		if err != nil {
			return nil, err
		}
	}

	return busMessages, nil
}
