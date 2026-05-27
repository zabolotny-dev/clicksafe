package attachmentapp

import "github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"

func attachmentContentType(atchType attachmentbus.AttachmentType) string {
	return atchType.MIMEType()
}
