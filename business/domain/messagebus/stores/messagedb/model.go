package messagedb

import (
	"net/mail"

	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus/stores/messagedb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

func toDBMessage(msg messagebus.Message) sqlc.Message {
	return sqlc.Message{
		ID:           msg.ID,
		Label:        msg.Label.String(),
		FromEmail:    msg.FromEmail.String(),
		FromName:     msg.FromName.ToSQLNullString(),
		Subject:      msg.Subject.ToSQLNullString(),
		ContentPath:  msg.ContentPath.ToSQLNullString(),
		RequiredVars: msg.RequiredVars,
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

	contentPath, err := file.ParseNull(msg.ContentPath.String)
	if err != nil {
		return messagebus.Message{}, err
	}

	return messagebus.Message{
		ID:           msg.ID,
		Label:        lbl,
		FromEmail:    *fromEmail,
		FromName:     fromName,
		Subject:      sub,
		ContentPath:  contentPath,
		RequiredVars: msg.RequiredVars,
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
