package attachmentdb

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus/stores/attachmentdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBAttachment(atch attachmentbus.Attachment) sqlc.Attachment {
	return sqlc.Attachment{
		ID:           atch.ID,
		Label:        atch.Label.String(),
		Type:         atch.Type.String(),
		ContentPath:  atch.ContentPath.String(),
		RequiredVars: atch.RequiredVars,
		IsPublic:     atch.Public,
		UploadedAt:   pgtype.Timestamptz{Time: atch.UploadedAt.UTC(), Valid: true},
	}
}

func toBusAttachment(atch sqlc.Attachment) (attachmentbus.Attachment, error) {
	lbl, err := label.Parse(atch.Label)
	if err != nil {
		return attachmentbus.Attachment{}, err
	}

	atchType, err := attachmentbus.Parse(atch.Type)
	if err != nil {
		return attachmentbus.Attachment{}, err
	}

	contentPath, err := file.Parse(atch.ContentPath)
	if err != nil {
		return attachmentbus.Attachment{}, err
	}

	if !atch.UploadedAt.Valid {
		return attachmentbus.Attachment{}, fmt.Errorf("uploaded_at cannot be null")
	}

	return attachmentbus.Attachment{
		ID:           atch.ID,
		Label:        lbl,
		Type:         atchType,
		ContentPath:  contentPath,
		RequiredVars: atch.RequiredVars,
		Public:       atch.IsPublic,
		UploadedAt:   atch.UploadedAt.Time.UTC(),
	}, nil
}

func toBusAttachments(attachments []sqlc.Attachment) ([]attachmentbus.Attachment, error) {
	busAttachments := make([]attachmentbus.Attachment, len(attachments))

	for i, atch := range attachments {
		var err error
		busAttachments[i], err = toBusAttachment(atch)
		if err != nil {
			return nil, err
		}
	}

	return busAttachments, nil
}
