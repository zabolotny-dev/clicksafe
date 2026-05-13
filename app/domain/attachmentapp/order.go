package attachmentapp

import "github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"

var orderByFields = map[string]string{
	"attachment_id": attachmentbus.OrderByID,
	"label":         attachmentbus.OrderByLabel,
	"type":          attachmentbus.OrderByType,
	"uploaded_at":   attachmentbus.OrderByUploadedAt,
}
