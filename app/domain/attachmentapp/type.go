package attachmentapp

import "github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"

func attachmentContentType(atchType attachmentbus.AttachmentType) string {
	switch atchType {
	case attachmentbus.Html:
		return "text/html"
	case attachmentbus.Docx:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case attachmentbus.Pptx:
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case attachmentbus.Xlsx:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case attachmentbus.Txt:
		return "text/plain"
	case attachmentbus.Png:
		return "image/png"
	case attachmentbus.Jpg, attachmentbus.Jpeg:
		return "image/jpeg"
	case attachmentbus.Gif:
		return "image/gif"
	case attachmentbus.Webp:
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
