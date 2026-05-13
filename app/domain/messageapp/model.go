package messageapp

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

type Message struct {
	ID           uuid.UUID     `json:"id"`
	Label        string        `json:"label"`
	FromEmail    string        `json:"from_email"`
	FromName     string        `json:"from_name"`
	Subject      string        `json:"subject"`
	AttachmentID uuid.NullUUID `json:"attachment_id"`
}

type NewMessage struct {
	Label        string `json:"label"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	Subject      string `json:"subject"`
	AttachmentID string `json:"attachment_id"`
}

type UpdateMessage struct {
	Label        *string `json:"label"`
	FromEmail    *string `json:"from_email"`
	FromName     *string `json:"from_name"`
	Subject      *string `json:"subject"`
	AttachmentID *string `json:"attachment_id"`
}

func toBusNewMessage(req NewMessage) (messagebus.NewMessage, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(req.Label)
	if err != nil {
		errors.Add("label", err)
	}

	fromEmail, err := mail.ParseAddress(req.FromEmail)
	if err != nil {
		errors.Add("from_email", err)
	}

	fromName, err := label.ParseNull(req.FromName)
	if err != nil {
		errors.Add("from_name", err)
	}

	sub, err := subject.ParseNull(req.Subject)
	if err != nil {
		errors.Add("subject", err)
	}

	var attachmentID uuid.NullUUID
	if req.AttachmentID != "" {
		parsed, err := uuid.Parse(req.AttachmentID)
		if err != nil {
			errors.Add("attachment_id", err)
		}
		attachmentID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	if len(errors) > 0 {
		return messagebus.NewMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return messagebus.NewMessage{
		Label:        lbl,
		FromEmail:    *fromEmail,
		FromName:     fromName,
		Subject:      sub,
		AttachmentID: attachmentID,
	}, nil
}

func toBusUpdateMessage(req UpdateMessage) (messagebus.UpdateMessage, error) {
	var errors errs.FieldErrors

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			errors.Add("label", err)
		}
		lbl = &parsed
	}

	var fromEmail *mail.Address
	if req.FromEmail != nil {
		parsed, err := mail.ParseAddress(*req.FromEmail)
		if err != nil {
			errors.Add("from_email", err)
		}
		fromEmail = parsed
	}

	var fromName *label.Null
	if req.FromName != nil {
		parsed, err := label.ParseNull(*req.FromName)
		if err != nil {
			errors.Add("from_name", err)
		}
		fromName = &parsed
	}

	var sub *subject.Null
	if req.Subject != nil {
		parsed, err := subject.ParseNull(*req.Subject)
		if err != nil {
			errors.Add("subject", err)
		}
		sub = &parsed
	}

	var attachmentID *uuid.NullUUID
	if req.AttachmentID != nil {
		attachmentID = &uuid.NullUUID{}
		if id := *req.AttachmentID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("attachment_id", err)
			}
			attachmentID.UUID = parsed
			attachmentID.Valid = err == nil
		}
	}

	if len(errors) > 0 {
		return messagebus.UpdateMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return messagebus.UpdateMessage{
		Label:        lbl,
		FromEmail:    fromEmail,
		FromName:     fromName,
		Subject:      sub,
		AttachmentID: attachmentID,
	}, nil
}

func toAppMessage(msg messagebus.Message) Message {
	return Message{
		ID:           msg.ID,
		Label:        msg.Label.String(),
		FromEmail:    msg.FromEmail.Address,
		FromName:     msg.FromName.String(),
		Subject:      msg.Subject.String(),
		AttachmentID: msg.AttachmentID,
	}
}

func toAppMessages(messages []messagebus.Message) []Message {
	items := make([]Message, len(messages))
	for i, msg := range messages {
		items[i] = toAppMessage(msg)
	}
	return items
}
