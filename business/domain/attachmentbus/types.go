package attachmentbus

import "fmt"

var (
	Docx = newType(".docx")
	Gif  = newType(".gif")
	Html = newType(".html")
	Jpeg = newType(".jpeg")
	Jpg  = newType(".jpg")
	Png  = newType(".png")
	Pptx = newType(".pptx")
	Txt  = newType(".txt")
	Webp = newType(".webp")
	Xlsx = newType(".xlsx")
)

var types = make(map[string]AttachmentType)

type AttachmentType struct {
	value string
}

func newType(value string) AttachmentType {
	e := AttachmentType{value}
	types[value] = e
	return e
}

func Parse(value string) (AttachmentType, error) {
	e, ok := types[value]
	if !ok {
		return AttachmentType{}, fmt.Errorf("invalid attachment type: '%s'", value)
	}

	return e, nil
}

func (e AttachmentType) String() string {
	return e.value
}

func (e AttachmentType) IsTemplate() bool {
	switch e {
	case Html, Docx, Pptx, Xlsx, Txt:
		return true
	default:
		return false
	}
}

func (e AttachmentType) IsMedia() bool {
	switch e {
	case Png, Jpg, Jpeg, Gif, Webp:
		return true
	default:
		return false
	}
}
