package attachmentapp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Attachment struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	Type         string    `json:"type"`
	RequiredVars []string  `json:"required_vars"`
	Public       bool      `json:"public"`
	PublicPath   *string   `json:"public_path,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type UpdateAttachment struct {
	Label  *string `json:"label"`
	Public *string `json:"public"`
}

func toBusNewAttachment(l string, public bool) (attachmentbus.NewAttachment, error) {
	var errors errs.FieldErrors

	lbl, err := label.Parse(l)
	if err != nil {
		errors.Add("label", err)
	}

	if len(errors) > 0 {
		return attachmentbus.NewAttachment{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return attachmentbus.NewAttachment{
		Label:  lbl,
		Public: public,
	}, nil
}

func toBusUpdateAttachment(req UpdateAttachment) (attachmentbus.UpdateAttachment, error) {
	var errors errs.FieldErrors

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			errors.Add("label", err)
		}
		lbl = &parsed
	}

	var publ *bool
	if req.Public != nil {
		parsed, err := strconv.ParseBool(*req.Public)
		if err != nil {
			errors.Add("public", err)
		}
		publ = &parsed
	}

	if len(errors) > 0 {
		return attachmentbus.UpdateAttachment{}, errors.ToError(errs.InvalidArgument, "validation failed")
	}

	return attachmentbus.UpdateAttachment{
		Label:  lbl,
		Public: publ,
	}, nil
}

func toAppAttachment(a attachmentbus.Attachment, p string) Attachment {
	if idx := strings.Index(p, "/:"); idx != -1 {
		p = p[:idx]
	}

	var publicPath *string
	if a.Public {
		path := fmt.Sprintf("%s/%s", p, a.ID.String())
		publicPath = &path
	}

	return Attachment{
		ID:           a.ID.String(),
		Label:        a.Label.String(),
		Type:         a.Type.String(),
		RequiredVars: a.RequiredVars,
		Public:       a.Public,
		PublicPath:   publicPath,
		UploadedAt:   a.UploadedAt,
	}
}

func toAppAttachments(attachments []attachmentbus.Attachment, path string) []Attachment {
	result := make([]Attachment, len(attachments))
	for i, a := range attachments {
		result[i] = toAppAttachment(a, path)
	}
	return result
}
