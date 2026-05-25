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
	ID            uuid.UUID     `json:"id"`
	Type          string        `json:"type"`
	Label         string        `json:"label"`
	FromEmail     string        `json:"from_email"`
	FromName      string        `json:"from_name"`
	Subject       string        `json:"subject"`
	HtmlBodyID    uuid.NullUUID `json:"html_body_id"`
	TextBodyID    uuid.NullUUID `json:"text_body_id"`
	MaxAccountID  uuid.NullUUID `json:"max_account_id"`
	AttachmentIDs []uuid.UUID   `json:"attachment_ids"`
}

type NewMessage struct {
	Type          string   `json:"type"`
	Label         string   `json:"label"`
	FromEmail     string   `json:"from_email"`
	FromName      string   `json:"from_name"`
	Subject       string   `json:"subject"`
	HtmlBodyID    string   `json:"html_body_id"`
	TextBodyID    string   `json:"text_body_id"`
	MaxAccountID  string   `json:"max_account_id"`
	AttachmentIDs []string `json:"attachment_ids"`
}

type UpdateMessage struct {
	Type          *string  `json:"type"`
	Label         *string  `json:"label"`
	FromEmail     *string  `json:"from_email"`
	FromName      *string  `json:"from_name"`
	Subject       *string  `json:"subject"`
	HtmlBodyID    *string  `json:"html_body_id"`
	TextBodyID    *string  `json:"text_body_id"`
	MaxAccountID  *string  `json:"max_account_id"`
	AttachmentIDs []string `json:"attachment_ids"`
}

func toBusNewMessage(req NewMessage) (messagebus.NewMessage, error) {
	var errors errs.FieldErrors

	msgType := messagebus.EmailMessage
	if req.Type != "" {
		parsed, err := messagebus.ParseMessageType(req.Type)
		if err != nil {
			errors.Add("type", err)
		} else {
			msgType = parsed
		}
	}

	lbl, err := label.Parse(req.Label)
	if err != nil {
		errors.Add("label", err)
	}

	var fromEmail *mail.Address
	if msgType == messagebus.EmailMessage || req.FromEmail != "" {
		parsed, err := mail.ParseAddress(req.FromEmail)
		if err != nil {
			errors.Add("from_email", err)
		} else {
			fromEmail = parsed
		}
	}

	fromName, err := label.ParseNull(req.FromName)
	if err != nil {
		errors.Add("from_name", err)
	}

	sub, err := subject.ParseNull(req.Subject)
	if err != nil {
		errors.Add("subject", err)
	}

	var htmlBodyID uuid.NullUUID
	if req.HtmlBodyID != "" {
		parsed, err := uuid.Parse(req.HtmlBodyID)
		if err != nil {
			errors.Add("html_body_id", err)
		}
		htmlBodyID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	var textBodyID uuid.NullUUID
	if req.TextBodyID != "" {
		parsed, err := uuid.Parse(req.TextBodyID)
		if err != nil {
			errors.Add("text_body_id", err)
		}
		textBodyID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	var maxAccountID uuid.NullUUID
	if req.MaxAccountID != "" {
		parsed, err := uuid.Parse(req.MaxAccountID)
		if err != nil {
			errors.Add("max_account_id", err)
		}
		maxAccountID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	var attachemntIDs []uuid.UUID
	if req.AttachmentIDs != nil {
		for _, id := range req.AttachmentIDs {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("attachment_ids", err)
			}
			attachemntIDs = append(attachemntIDs, parsed)
		}
	}

	if len(errors) > 0 {
		return messagebus.NewMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	var address mail.Address
	if fromEmail != nil {
		address = *fromEmail
	}

	return messagebus.NewMessage{
		Type:          msgType,
		Label:         lbl,
		FromEmail:     address,
		FromName:      fromName,
		Subject:       sub,
		HtmlBodyID:    htmlBodyID,
		TextBodyID:    textBodyID,
		MaxAccountID:  maxAccountID,
		AttachmentIDs: attachemntIDs,
	}, nil
}

func toBusUpdateMessage(req UpdateMessage) (messagebus.UpdateMessage, error) {
	var errors errs.FieldErrors

	var lbl *label.Label
	var msgType *messagebus.MessageType
	if req.Type != nil {
		parsed, err := messagebus.ParseMessageType(*req.Type)
		if err != nil {
			errors.Add("type", err)
		} else {
			msgType = &parsed
		}
	}

	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			errors.Add("label", err)
		}
		lbl = &parsed
	}

	var fromEmail *mail.Address
	if req.FromEmail != nil {
		if *req.FromEmail == "" {
			fromEmail = &mail.Address{}
		} else {
			parsed, err := mail.ParseAddress(*req.FromEmail)
			if err != nil {
				errors.Add("from_email", err)
			}
			fromEmail = parsed
		}
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

	var htmlBodyID *uuid.NullUUID
	if req.HtmlBodyID != nil {
		htmlBodyID = &uuid.NullUUID{}
		if id := *req.HtmlBodyID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("html_body_id", err)
			}
			htmlBodyID.UUID = parsed
			htmlBodyID.Valid = err == nil
		}
	}

	var textBodyID *uuid.NullUUID
	if req.TextBodyID != nil {
		textBodyID = &uuid.NullUUID{}
		if id := *req.TextBodyID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("text_body_id", err)
			}
			textBodyID.UUID = parsed
			textBodyID.Valid = err == nil
		}
	}

	var maxAccountID *uuid.NullUUID
	if req.MaxAccountID != nil {
		maxAccountID = &uuid.NullUUID{}
		if id := *req.MaxAccountID; id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("max_account_id", err)
			}
			maxAccountID.UUID = parsed
			maxAccountID.Valid = err == nil
		}
	}

	var attachemntIDs []uuid.UUID
	if req.AttachmentIDs != nil {
		for _, id := range req.AttachmentIDs {
			parsed, err := uuid.Parse(id)
			if err != nil {
				errors.Add("attachment_ids", err)
			}
			attachemntIDs = append(attachemntIDs, parsed)
		}
	}

	if len(errors) > 0 {
		return messagebus.UpdateMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return messagebus.UpdateMessage{
		Type:          msgType,
		Label:         lbl,
		FromEmail:     fromEmail,
		FromName:      fromName,
		Subject:       sub,
		HtmlBodyID:    htmlBodyID,
		TextBodyID:    textBodyID,
		MaxAccountID:  maxAccountID,
		AttachmentIDs: attachemntIDs,
	}, nil
}

func toAppMessage(msg messagebus.Message) Message {
	return Message{
		ID:            msg.ID,
		Type:          msg.Type.String(),
		Label:         msg.Label.String(),
		FromEmail:     msg.FromEmail.Address,
		FromName:      msg.FromName.String(),
		Subject:       msg.Subject.String(),
		HtmlBodyID:    msg.HtmlBodyID,
		TextBodyID:    msg.TextBodyID,
		MaxAccountID:  msg.MaxAccountID,
		AttachmentIDs: msg.AttachmentIDs,
	}
}

func toAppMessages(messages []messagebus.Message) []Message {
	items := make([]Message, len(messages))
	for i, msg := range messages {
		items[i] = toAppMessage(msg)
	}
	return items
}
