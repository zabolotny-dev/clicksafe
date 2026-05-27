package attachmentbus

import "fmt"

var (
	Aac  = newType(".aac")
	Docx = newType(".docx")
	Flac = newType(".flac")
	Gif  = newType(".gif")
	Html = newType(".html")
	Jpeg = newType(".jpeg")
	Jpg  = newType(".jpg")
	M4a  = newType(".m4a")
	Mp3  = newType(".mp3")
	Ogg  = newType(".ogg")
	Opus = newType(".opus")
	Png  = newType(".png")
	Pptx = newType(".pptx")
	Txt  = newType(".txt")
	Wav  = newType(".wav")
	Weba = newType(".weba")
	Webm = newType(".webm")
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
	return e.IsImage() || e.IsAudio()
}

func (e AttachmentType) IsImage() bool {
	switch e {
	case Png, Jpg, Jpeg, Gif, Webp:
		return true
	default:
		return false
	}
}

func (e AttachmentType) IsAudio() bool {
	switch e {
	case Aac, Flac, M4a, Mp3, Ogg, Opus, Wav, Weba, Webm:
		return true
	default:
		return false
	}
}

func (e AttachmentType) MIMEType() string {
	switch e {
	case Aac:
		return "audio/aac"
	case Docx:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case Flac:
		return "audio/flac"
	case Gif:
		return "image/gif"
	case Html:
		return "text/html"
	case Jpeg, Jpg:
		return "image/jpeg"
	case M4a:
		return "audio/mp4"
	case Mp3:
		return "audio/mpeg"
	case Ogg:
		return "audio/ogg"
	case Opus:
		return "audio/ogg; codecs=opus"
	case Png:
		return "image/png"
	case Pptx:
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case Txt:
		return "text/plain"
	case Wav:
		return "audio/wav"
	case Weba, Webm:
		return "audio/webm"
	case Webp:
		return "image/webp"
	case Xlsx:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
