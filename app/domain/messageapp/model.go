package messageapp

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

type Message struct {
	ID           uuid.UUID `json:"id"`
	Label        string    `json:"label"`
	FromEmail    string    `json:"from_email"`
	FromName     string    `json:"from_name"`
	Subject      string    `json:"subject"`
	HasContent   bool      `json:"has_content"`
	RequiredVars []string  `json:"required_vars"`
}

type NewMessage struct {
	Label     string `json:"label"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	Subject   string `json:"subject"`
}

type UpdateMessage struct {
	Label     *string `json:"label"`
	FromEmail *string `json:"from_email"`
	FromName  *string `json:"from_name"`
	Subject   *string `json:"subject"`
}

type RenderMessage struct {
	EmployeeID string `json:"employee_id"`
}

type RenderedMessage struct {
	Content string `json:"content"`
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

	if len(errors) > 0 {
		return messagebus.NewMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return messagebus.NewMessage{
		Label:     lbl,
		FromEmail: *fromEmail,
		FromName:  fromName,
		Subject:   sub,
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

	if len(errors) > 0 {
		return messagebus.UpdateMessage{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return messagebus.UpdateMessage{
		Label:     lbl,
		FromEmail: fromEmail,
		FromName:  fromName,
		Subject:   sub,
	}, nil
}

func toBusRenderScope(req RenderMessage) (resolverbus.Scope, error) {
	var errors errs.FieldErrors
	var scope resolverbus.Scope

	if req.EmployeeID != "" {
		id, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			errors.Add("employee_id", err)
		}
		scope.EmployeeID = id
	}

	if len(errors) > 0 {
		return resolverbus.Scope{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return scope, nil
}

func toAppMessage(msg messagebus.Message) Message {
	return Message{
		ID:           msg.ID,
		Label:        msg.Label.String(),
		FromEmail:    msg.FromEmail.Address,
		FromName:     msg.FromName.String(),
		Subject:      msg.Subject.String(),
		HasContent:   msg.ContentPath.Valid(),
		RequiredVars: msg.RequiredVars,
	}
}

func toAppMessages(messages []messagebus.Message) []Message {
	items := make([]Message, len(messages))
	for i, msg := range messages {
		items[i] = toAppMessage(msg)
	}
	return items
}
